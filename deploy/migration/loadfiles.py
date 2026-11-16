"""Render the pgloader load files for the MySQL -> Postgres copy (#569).

Three data-only passes: messages first (email_reminders references it), then
the rest of the app tables, then the Slack OAuth tables. The target schema is
created beforehand by `app database migrate up` and the slackbot's own
create_all, so pgloader must not touch DDL or constraints:
  * `data only`        - no CREATE/DROP of tables or indexes
  * `no foreign keys`  - otherwise pgloader drops the target FKs and does not
                         recreate them in data-only mode
  * `ALTER SCHEMA ... RENAME TO 'public'` - pgloader maps a MySQL database to
                         a Postgres schema of the same name; ours live in public
MySQL is read with the server's own time zone, which the runbook gates to UTC.
schema_migrations and slack_oauth_states are never copied.

Usage: python -I loadfiles.py OUTDIR [--suffix _rehearsal]  (creds from env)
"""
from __future__ import annotations

import argparse
import os
import sys
from pathlib import Path
from urllib.parse import quote

_COMMON = """WITH data only, no foreign keys, workers = 2, concurrency = 1
SET PostgreSQL PARAMETERS timezone = 'UTC'
"""
_APP_CAST = """CAST type tinyint when (= precision 1) to boolean using tinyint-to-boolean,
     type timestamp to timestamptz using zero-dates-to-null,
     type datetime to timestamptz using zero-dates-to-null
"""
_SLACK_CAST = """CAST type tinyint when (= precision 1) to boolean using tinyint-to-boolean,
     type datetime to timestamp using zero-dates-to-null
"""


def _q(s: str) -> str:
    return quote(s, safe="")


def render(env: dict, suffix: str, host: str = "127.0.0.1",
           my_port: int = 3306, pg_port: int = 5432) -> dict:
    """Return {filename: load-file text}, in load order."""
    my_app = f"mysql://{_q(env['MY_APP_USER'])}:{_q(env['MY_APP_PASS'])}@{host}:{my_port}/{env['MY_APP_DB']}"
    my_oauth = f"mysql://{_q(env['MY_OAUTH_USER'])}:{_q(env['MY_OAUTH_PASS'])}@{host}:{my_port}/{env['MY_OAUTH_DB']}"
    pg_app = f"postgresql://passwordexchange:{_q(env['PG_APP_PASS'])}@{host}:{pg_port}/passwordexchange{suffix}?sslmode=disable"
    pg_slack = f"postgresql://slackbot:{_q(env['PG_SLACK_PASS'])}@{host}:{pg_port}/slack_oauth{suffix}?sslmode=disable"
    passes = [
        ("1-messages.load", my_app, env["MY_APP_DB"], pg_app, _APP_CAST, ["messages"]),
        ("2-app-rest.load", my_app, env["MY_APP_DB"], pg_app, _APP_CAST,
         ["email_reminders", "file_upload_sessions"]),
        ("3-slack.load", my_oauth, env["MY_OAUTH_DB"], pg_slack, _SLACK_CAST,
         ["slack_installations", "slack_bots"]),
    ]
    out = {}
    for name, src, src_db, dst, cast, tables in passes:
        matching = ", ".join(f"'{t}'" for t in tables)
        out[name] = (
            f"LOAD DATABASE FROM {src}\n     INTO {dst}\n{_COMMON}{cast}"
            f"INCLUDING ONLY TABLE NAMES MATCHING {matching}\n"
            f"ALTER SCHEMA '{src_db}' RENAME TO 'public';\n"
        )
    return out


def write(files: dict, outdir) -> None:
    """Write load files readable by the owner only (they embed passwords)."""
    for name, body in files.items():
        fd = os.open(Path(outdir) / name, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
        try:
            os.write(fd, body.encode())
        finally:
            os.close(fd)
        os.chmod(Path(outdir) / name, 0o600)


def main(argv=None) -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("outdir")
    ap.add_argument("--suffix", default="")
    args = ap.parse_args(argv)
    files = render(os.environ, args.suffix)
    write(files, args.outdir)
    print("wrote", " ".join(files))
    return 0


if __name__ == "__main__":
    sys.exit(main())
