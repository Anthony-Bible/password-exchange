"""Shared fixtures for the slackbot test-suite.

Container-backed tests need a container runtime. Under podman (instead of
Docker) point testcontainers at the podman socket and disable the Ryuk reaper,
same as the Go harness in app/internal/integration/dbtest/postgres.go:

    # macOS
    export DOCKER_HOST=unix://$(podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}')
    export TESTCONTAINERS_RYUK_DISABLED=true
    pytest slackbot/tests

Images can be overridden with PASSWORDEXCHANGE_TEST_PG_IMAGE and
PASSWORDEXCHANGE_TEST_MYSQL_IMAGE.
"""
import os
import sys
import time
from pathlib import Path

import pytest

TESTS_DIR = Path(__file__).resolve().parent
SLACKBOT_DIR = TESTS_DIR.parent
REPO_ROOT = SLACKBOT_DIR.parent

# slack-sdk stores naive datetimes and reads them back in local time, so
# installed_at only round-trips exactly when the process runs in UTC, as the
# container does. Pin it so developer machines behave like CI.
os.environ["TZ"] = "UTC"
time.tzset()

# Make `import oauth_db` work without installing the slackbot as a package.
if str(SLACKBOT_DIR) not in sys.path:
    sys.path.insert(0, str(SLACKBOT_DIR))

PG_IMAGE = os.environ.get(
    "PASSWORDEXCHANGE_TEST_PG_IMAGE", "mirror.gcr.io/library/postgres:17-alpine"
)
MYSQL_IMAGE = os.environ.get(
    "PASSWORDEXCHANGE_TEST_MYSQL_IMAGE", "mirror.gcr.io/library/mysql:8.0"
)

DB_USER = "oauth"
DB_PASSWORD = "p@ss:w/rd?#%"  # deliberately URL-hostile
DB_NAME = "oauthdb"


def _ip(host: str) -> str:
    # MySQLdb treats "localhost" as a unix socket and ignores the port.
    return "127.0.0.1" if host == "localhost" else host


@pytest.fixture(scope="session")
def postgres_env():
    """Env mapping (as the bot would see it) for a throwaway Postgres 17."""
    from testcontainers.postgres import PostgresContainer

    with PostgresContainer(
        PG_IMAGE, username=DB_USER, password=DB_PASSWORD, dbname=DB_NAME, driver=None
    ) as pg:
        yield {
            "PASSWORDEXCHANGE_DBDRIVER": "postgres",
            "OAUTHDB_USER": DB_USER,
            "OAUTHDB_PASSWORD": DB_PASSWORD,
            "OAUTHDB_NAME": DB_NAME,
            "PASSWORDEXCHANGE_DBHOST": _ip(pg.get_container_host_ip()),
            "PASSWORDEXCHANGE_DBPORT": str(pg.get_exposed_port(5432)),
        }


@pytest.fixture(scope="session")
def mysql_env():
    """Env mapping for a throwaway MySQL 8.0. Host is given as host:port, the
    way existing deployments may configure PASSWORDEXCHANGE_DBHOST."""
    from testcontainers.mysql import MySqlContainer

    with MySqlContainer(
        MYSQL_IMAGE, username=DB_USER, password=DB_PASSWORD, dbname=DB_NAME
    ) as my:
        host = _ip(my.get_container_host_ip())
        yield {
            "PASSWORDEXCHANGE_DBDRIVER": "mysql",
            "OAUTHDB_USER": DB_USER,
            "OAUTHDB_PASSWORD": DB_PASSWORD,
            "OAUTHDB_NAME": DB_NAME,
            "PASSWORDEXCHANGE_DBHOST": f"{host}:{my.get_exposed_port(3306)}",
        }


@pytest.fixture(scope="session")
def protos_shim(tmp_path_factory):
    """Directory exposing python_protos/ as the `protos` package, mirroring the
    Dockerfile's `COPY python_protos ./slackbot/protos`."""
    d = tmp_path_factory.mktemp("protos_shim")
    (d / "protos").symlink_to(REPO_ROOT / "python_protos")
    return d


@pytest.fixture
def program_env(protos_shim):
    """Minimal environment that lets slackbot/program.py import (no Slack calls
    are made at import time). DB settings are supplied by each test."""
    env = {k: v for k, v in os.environ.items() if not k.startswith(("OAUTHDB_", "PASSWORDEXCHANGE_"))}
    env.update(
        {
            "PYTHONPATH": os.pathsep.join([str(SLACKBOT_DIR), str(protos_shim), str(REPO_ROOT / "python_protos")]),
            "PASSWORDEXCHANGE_DATABASESERVICE": "localhost:1",
            "PASSWORDEXCHANGE_ENCRYPTIONSERVICE": "localhost",
            "PASSWORDEXCHANGE_HOST": "https://example.invalid/",
            "SLACK_CLIENT_ID": "111.222",
            "SLACK_CLIENT_SECRET": "secret",
            "SLACK_SIGNING_SECRET": "signing",
        }
    )
    return env
