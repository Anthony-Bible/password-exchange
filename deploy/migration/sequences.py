"""Point the Postgres identity sequences past every id MySQL ever issued (#569).

The target is MySQL's AUTO_INCREMENT, not MAX(id): ids of deleted rows are
never handed out again. AUTO_INCREMENT is read with information_schema stats
caching disabled (the server caches it for a day by default).

Usage: python -I sequences.py [--suffix _rehearsal]  (creds from env)
"""
from __future__ import annotations

import argparse
import os
import sys

# (mysql db env var, postgres db, postgres user, password env var, {table: id column})
TARGETS = [
    ("MY_APP", "passwordexchange", "passwordexchange", "PG_APP_PASS",
     {"messages": "messageid", "email_reminders": "id"}),
    ("MY_OAUTH", "slack_oauth", "slackbot", "PG_SLACK_PASS",
     {"slack_installations": "id", "slack_bots": "id"}),
]


def setval_args(auto_increment: int, src_max) -> tuple:
    """(value, is_called) so the next nextval() is max(AUTO_INCREMENT, MAX(id)+1)."""
    next_id = max(auto_increment, (src_max or 0) + 1)
    return (next_id - 1, True) if next_id > 1 else (1, False)


def main(argv=None) -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--suffix", default="")
    ap.add_argument("--host", default="127.0.0.1")
    ap.add_argument("--my-port", type=int, default=13306)
    ap.add_argument("--pg-port", type=int, default=15432)
    args = ap.parse_args(argv)

    import MySQLdb
    import psycopg

    e = os.environ
    for my, pg_db, pg_user, pg_pw, tables in TARGETS:
        mc = MySQLdb.connect(host=args.host, port=args.my_port, user=e[f"{my}_USER"],
                             passwd=e[f"{my}_PASS"], db=e[f"{my}_DB"]).cursor()
        mc.execute("SET SESSION information_schema_stats_expiry = 0")
        with psycopg.connect(host=args.host, port=args.pg_port, user=pg_user, password=e[pg_pw],
                             dbname=pg_db + args.suffix, sslmode="disable") as pg:
            for table, col in tables.items():
                mc.execute("SELECT AUTO_INCREMENT FROM information_schema.tables "
                           "WHERE table_schema = DATABASE() AND table_name = %s", (table,))
                auto_increment = mc.fetchone()[0]
                mc.execute(f"SELECT MAX({col}) FROM {table}")
                src_max = mc.fetchone()[0]
                value, called = setval_args(auto_increment, src_max)
                pg.execute(f"SELECT setval(pg_get_serial_sequence('{table}', '{col}'), %s, %s)",
                           (value, called))
                print(f"{table}.{col}: AUTO_INCREMENT {auto_increment}, max {src_max} -> next id {value + 1 if called else value}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
