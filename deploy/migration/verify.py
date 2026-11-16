"""Row-by-row verification of the MySQL -> Postgres copy (#569).

Reads every migrated table from both sides, normalizes representation-only
differences (JSON formatting, tinyint vs bool, hex case, timezone-aware vs naive
UTC timestamps) and reports missing, extra and differing rows per table, plus
whether each Postgres sequence will hand out an id above the source max.

The report prints counts, column names and digests only. It never prints row
values or keys: the tables hold message ciphertext, emails and Slack tokens.

Usage (credentials come from the environment, see creds.env in the runbook):
    python -I verify.py --suffix _rehearsal [--my-port 13306 --pg-port 15432]
Exit status: 0 = everything matches, 1 = a mismatch, 2 = usage/connection error.
"""
from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import os
import sys
from dataclasses import dataclass, field

NULL = ("<null>",)  # sentinel, distinct from every real normalized value


def normalize(kind: str, value):
    """Map a cell from either database to a comparable value."""
    if value is None:
        return NULL
    if kind == "bool":
        return int(bool(value))
    if kind == "int":
        return int(value)
    if kind == "hex":
        return (value.decode() if isinstance(value, bytes) else str(value)).lower()
    if kind == "json":
        text = value.decode() if isinstance(value, bytes) else value
        return json.dumps(json.loads(text), sort_keys=True, separators=(",", ":"))
    if kind == "ts":
        if isinstance(value, dt.datetime):
            if value.tzinfo is not None:
                value = value.astimezone(dt.timezone.utc).replace(tzinfo=None)
            return value.isoformat()
        return str(value)
    if isinstance(value, bytes):
        return value.decode()
    return str(value)


@dataclass
class Report:
    src_rows: int
    dst_rows: int
    missing: list = field(default_factory=list)
    extra: list = field(default_factory=list)
    differing: dict = field(default_factory=dict)

    @property
    def ok(self) -> bool:
        return not (self.missing or self.extra or self.differing)

    def render(self, table: str) -> str:
        if self.ok:
            return f"{table}: OK ({self.src_rows} rows)"
        cols = sorted({c for cs in self.differing.values() for c in cs})
        return (
            f"{table}: MISMATCH source {self.src_rows} rows / target {self.dst_rows} rows; "
            f"{len(self.missing)} missing; {len(self.extra)} extra; "
            f"{len(self.differing)} differing in columns [{' '.join(cols)}]"
        )


def compare(src: dict, dst: dict, columns: list) -> Report:
    """Compare two {key: tuple-of-normalized-cells} maps."""
    report = Report(len(src), len(dst))
    report.missing = sorted(k for k in src if k not in dst)
    report.extra = sorted(k for k in dst if k not in src)
    for key in src.keys() & dst.keys():
        cols = [c for c, a, b in zip(columns, src[key], dst[key]) if a != b]
        if cols:
            report.differing[key] = cols
    return report


def digest(table: dict) -> str:
    """Order-independent SHA-256 over a normalized table."""
    h = hashlib.sha256()
    for key in sorted(table, key=repr):
        h.update(repr((key, table[key])).encode())
        h.update(b"\n")
    return h.hexdigest()


def next_id_ok(last_value: int, is_called: bool, src_max) -> bool:
    """True when nextval() will return an id above every copied id."""
    next_id = last_value + 1 if is_called else last_value
    return next_id > (src_max or 0)


# --- table specs: (column, kind); the first column is the comparison key ---

APP_TABLES = {
    "messages": [
        ("messageid", "int"), ("created", "ts"), ("lastaccessed", "ts"), ("viewed", "int"),
        ("firstname", "str"), ("lastname", "str"), ("other_firstname", "str"),
        ("other_lastname", "str"), ("message", "str"), ("email", "str"), ("other_email", "str"),
        ("uniqueid", "str"), ("is_client_encrypted", "bool"), ("view_count", "int"),
        ("max_view_count", "int"), ("expires_at", "ts"),
    ],
    "email_reminders": [
        ("id", "int"), ("message_id", "int"), ("email_address", "str"), ("reminder_count", "int"),
        ("last_reminder_sent", "ts"), ("created_at", "ts"),
    ],
    "file_upload_sessions": [
        ("session_id", "str"), ("file_id", "str"), ("message_id", "str"), ("upload_id", "str"),
        ("filename", "str"), ("content_type", "str"), ("total_size", "int"),
        ("total_chunks", "int"), ("status", "str"), ("encryption_key", "hex"),
        ("completed_parts", "json"), ("created_at", "ts"), ("expires_at", "ts"),
    ],
}
_SLACK_COMMON = [
    ("id", "int"), ("client_id", "str"), ("app_id", "str"), ("enterprise_id", "str"),
    ("enterprise_name", "str"),
]
SLACK_TABLES = {
    "slack_bots": _SLACK_COMMON + [
        ("team_id", "str"), ("team_name", "str"), ("bot_token", "str"), ("bot_id", "str"),
        ("bot_user_id", "str"), ("bot_scopes", "str"), ("bot_refresh_token", "str"),
        ("bot_token_expires_at", "ts"), ("is_enterprise_install", "bool"), ("installed_at", "ts"),
    ],
    "slack_installations": _SLACK_COMMON + [
        ("enterprise_url", "str"), ("team_id", "str"), ("team_name", "str"), ("bot_token", "str"),
        ("bot_id", "str"), ("bot_user_id", "str"), ("bot_scopes", "str"),
        ("bot_refresh_token", "str"), ("bot_token_expires_at", "ts"), ("user_id", "str"),
        ("user_token", "str"), ("user_scopes", "str"), ("user_refresh_token", "str"),
        ("user_token_expires_at", "ts"), ("incoming_webhook_url", "str"),
        ("incoming_webhook_channel", "str"), ("incoming_webhook_channel_id", "str"),
        ("incoming_webhook_configuration_url", "str"), ("is_enterprise_install", "bool"),
        ("token_type", "str"), ("installed_at", "ts"),
    ],
}
SEQUENCES = {"messages": "messageid", "email_reminders": "id", "slack_bots": "id",
             "slack_installations": "id"}


def _select(dialect: str, table: str, spec) -> str:
    exprs = []
    for col, kind in spec:
        if kind == "hex":
            exprs.append(f"HEX({col})" if dialect == "mysql" else f"encode({col}, 'hex')")
        elif kind == "json" and dialect == "pg":
            exprs.append(f"{col}::text")
        else:
            exprs.append(col)
    return f"SELECT {', '.join(exprs)} FROM {table}"


def _load(cur, dialect, table, spec) -> dict:
    cur.execute(_select(dialect, table, spec))
    kinds = [k for _, k in spec]
    out = {}
    for row in cur.fetchall():
        cells = tuple(normalize(k, v) for k, v in zip(kinds, row))
        out[cells[0]] = cells[1:]
    return out


def _verify_db(my, pg, tables) -> bool:
    ok = True
    mc, pc = my.cursor(), pg.cursor()
    for table, spec in tables.items():
        src, dst = _load(mc, "mysql", table, spec), _load(pc, "pg", table, spec)
        report = compare(src, dst, [c for c, _ in spec[1:]])
        same = digest(src) == digest(dst)
        print(f"{report.render(table)}; digest {'match' if same else 'DIFFER'} {digest(dst)[:16]}")
        ok = ok and report.ok and same
        if table in SEQUENCES:
            col = SEQUENCES[table]
            mc.execute(f"SELECT MAX({col}) FROM {table}")
            src_max = mc.fetchone()[0]
            pc.execute(f"SELECT pg_get_serial_sequence('{table}', '{col}')")
            seq = pc.fetchone()[0]
            pc.execute(f"SELECT last_value, is_called FROM {seq}")
            last, called = pc.fetchone()
            seq_ok = next_id_ok(last, called, src_max)
            print(f"  sequence {seq}: next id {'above' if seq_ok else 'NOT above'} source max {src_max}")
            ok = ok and seq_ok
    return ok


def main(argv=None) -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--suffix", default="", help="target DB suffix, e.g. _rehearsal")
    ap.add_argument("--host", default="127.0.0.1")
    ap.add_argument("--my-port", type=int, default=13306)
    ap.add_argument("--pg-port", type=int, default=15432)
    args = ap.parse_args(argv)

    import MySQLdb
    import psycopg

    e = os.environ

    def mysql(user, pw, db):
        conn = MySQLdb.connect(host=args.host, port=args.my_port, user=e[user], passwd=e[pw],
                               db=e[db], charset="utf8mb4")
        conn.cursor().execute("SET time_zone = '+00:00'")
        return conn

    def pg(user, pw, db):
        conn = psycopg.connect(host=args.host, port=args.pg_port, user=user, password=e[pw],
                               dbname=db + args.suffix, sslmode="disable")
        conn.execute("SET TIME ZONE 'UTC'")
        return conn

    try:
        pairs = [
            (mysql("MY_APP_USER", "MY_APP_PASS", "MY_APP_DB"),
             pg("passwordexchange", "PG_APP_PASS", "passwordexchange"), APP_TABLES),
            (mysql("MY_OAUTH_USER", "MY_OAUTH_PASS", "MY_OAUTH_DB"),
             pg("slackbot", "PG_SLACK_PASS", "slack_oauth"), SLACK_TABLES),
        ]
    except Exception as exc:  # noqa: BLE001 - connection errors carry no secrets
        print(f"connection failed: {type(exc).__name__}", file=sys.stderr)
        return 2
    ok = all([_verify_db(my, p, tables) for my, p, tables in pairs])
    print("RESULT:", "PASS" if ok else "FAIL")
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
