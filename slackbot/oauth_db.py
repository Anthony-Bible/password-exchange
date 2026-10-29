"""Engine construction for the Slack OAuth installation/state stores.

The backing database is selected by PASSWORDEXCHANGE_DBDRIVER ("mysql", the
default, or "postgres"); credentials come from the OAUTHDB_* variables.
"""
from sqlalchemy import create_engine
from sqlalchemy.engine import URL, Engine

DEFAULT_DRIVER = "mysql"
DEFAULT_POSTGRES_PORT = 5432


def _port(raw, variable) -> int:
    try:
        return int(raw)
    except ValueError:
        raise ValueError("invalid port %r in %s (expected an integer)" % (raw, variable)) from None


def _postgres_url(env, **credentials) -> URL:
    return URL.create(
        "postgresql+psycopg",
        host=env.get("PASSWORDEXCHANGE_DBHOST"),
        port=_port(env.get("PASSWORDEXCHANGE_DBPORT") or DEFAULT_POSTGRES_PORT, "PASSWORDEXCHANGE_DBPORT"),
        **credentials,
    )


def _mysql_url(env, **credentials) -> URL:
    # Existing deployments may configure PASSWORDEXCHANGE_DBHOST as "host:port".
    host, port = env.get("PASSWORDEXCHANGE_DBHOST"), None
    if host and ":" in host:
        host, raw_port = host.rsplit(":", 1)
        port = _port(raw_port, "PASSWORDEXCHANGE_DBHOST")
    return URL.create("mysql+mysqldb", host=host, port=port, **credentials)


_URL_BUILDERS = {"postgres": _postgres_url, "mysql": _mysql_url}


def build_url(env) -> URL:
    driver = env.get("PASSWORDEXCHANGE_DBDRIVER") or DEFAULT_DRIVER
    try:
        builder = _URL_BUILDERS[driver]
    except KeyError:
        raise ValueError(
            "unsupported PASSWORDEXCHANGE_DBDRIVER %r (expected 'postgres' or 'mysql')" % driver
        ) from None
    return builder(
        env,
        username=env.get("OAUTHDB_USER"),
        password=env.get("OAUTHDB_PASSWORD"),
        database=env.get("OAUTHDB_NAME"),
    )


def make_engine(env) -> Engine:
    return create_engine(
        build_url(env),
        pool_pre_ping=True,
        pool_size=2,
        max_overflow=2,
        pool_recycle=1800,
    )
