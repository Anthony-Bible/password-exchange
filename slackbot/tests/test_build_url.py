"""Contract of oauth_db.build_url / make_engine (no database needed)."""
import pytest
from hypothesis import given, settings, strategies as st
from sqlalchemy.engine import URL, make_url

import oauth_db

# No NUL bytes, no lone surrogates; everything else (incl. @:/?#% and spaces).
text = st.text(
    alphabet=st.characters(blacklist_categories=("Cs",), blacklist_characters="\x00"),
    min_size=1,
    max_size=40,
)
hostname = st.from_regex(r"[a-z0-9]([a-z0-9.-]{0,30}[a-z0-9])?", fullmatch=True)
port = st.integers(min_value=1, max_value=65535)
# SQLAlchemy does not percent-encode the database segment when rendering a URL
# string, so "?" (the query delimiter) cannot survive a *string* round-trip.
# That is a SQLAlchemy limitation, not a build_url defect; the URL object itself
# must still carry any database name verbatim (see the *_object_* tests).
db_text = st.text(
    alphabet=st.characters(blacklist_categories=("Cs",), blacklist_characters="\x00?"),
    min_size=1,
    max_size=40,
)


def roundtrip(url: URL) -> URL:
    # SQLAlchemy 2.0 masks the password in str(url); render explicitly.
    return make_url(url.render_as_string(hide_password=False))


@settings(max_examples=200)
@given(user=text, password=text, db=db_text, host=hostname, p=port)
def test_postgres_url_roundtrips_arbitrary_credentials(user, password, db, host, p):
    env = {
        "PASSWORDEXCHANGE_DBDRIVER": "postgres",
        "OAUTHDB_USER": user,
        "OAUTHDB_PASSWORD": password,
        "OAUTHDB_NAME": db,
        "PASSWORDEXCHANGE_DBHOST": host,
        "PASSWORDEXCHANGE_DBPORT": str(p),
    }
    url = roundtrip(oauth_db.build_url(env))
    assert (url.username, url.password, url.host, url.port, url.database) == (
        user, password, host, p, db,
    )


@settings(max_examples=200)
@given(user=text, password=text, db=db_text, host=hostname, p=port)
def test_mysql_url_roundtrips_arbitrary_credentials_with_host_port(user, password, db, host, p):
    env = {
        "PASSWORDEXCHANGE_DBDRIVER": "mysql",
        "OAUTHDB_USER": user,
        "OAUTHDB_PASSWORD": password,
        "OAUTHDB_NAME": db,
        "PASSWORDEXCHANGE_DBHOST": f"{host}:{p}",
    }
    url = roundtrip(oauth_db.build_url(env))
    assert (url.username, url.password, url.host, url.port, url.database) == (
        user, password, host, p, db,
    )


BASE = {
    "OAUTHDB_USER": "u",
    "OAUTHDB_PASSWORD": "pw",
    "OAUTHDB_NAME": "d",
    "PASSWORDEXCHANGE_DBHOST": "db.internal",
}


def test_default_driver_is_mysql_via_mysqldb():
    url = oauth_db.build_url(dict(BASE))
    assert url.drivername == "mysql+mysqldb"
    assert (url.host, url.port) == ("db.internal", None)


def test_mysql_host_without_port_keeps_port_unset():
    url = oauth_db.build_url({**BASE, "PASSWORDEXCHANGE_DBDRIVER": "mysql"})
    assert url.drivername == "mysql+mysqldb"
    assert url.port is None


def test_mysql_host_with_port_is_split():
    url = oauth_db.build_url({**BASE, "PASSWORDEXCHANGE_DBHOST": "db.internal:3307"})
    assert (url.host, url.port) == ("db.internal", 3307)


def test_postgres_uses_psycopg3_and_defaults_port_5432():
    url = oauth_db.build_url({**BASE, "PASSWORDEXCHANGE_DBDRIVER": "postgres"})
    assert url.drivername == "postgresql+psycopg"
    assert url.port == 5432


@pytest.mark.parametrize("raw", ["", None])
def test_postgres_blank_port_falls_back_to_5432(raw):
    env = {**BASE, "PASSWORDEXCHANGE_DBDRIVER": "postgres"}
    if raw is not None:
        env["PASSWORDEXCHANGE_DBPORT"] = raw
    assert oauth_db.build_url(env).port == 5432


def test_url_is_a_sqlalchemy_url_object_not_a_string():
    assert isinstance(oauth_db.build_url(dict(BASE)), URL)


@pytest.mark.parametrize("fn", ["build_url", "make_engine"])
def test_unknown_driver_fails_naming_the_variable(fn):
    with pytest.raises(Exception, match="PASSWORDEXCHANGE_DBDRIVER"):
        getattr(oauth_db, fn)({**BASE, "PASSWORDEXCHANGE_DBDRIVER": "bogus"})


@pytest.mark.parametrize(
    "driver, overrides, variable",
    [
        ("postgres", {"PASSWORDEXCHANGE_DBPORT": "not-a-port"}, "PASSWORDEXCHANGE_DBPORT"),
        ("mysql", {"PASSWORDEXCHANGE_DBHOST": "db.internal:not-a-port"}, "PASSWORDEXCHANGE_DBHOST"),
    ],
)
def test_non_numeric_port_fails_naming_the_variable(driver, overrides, variable):
    with pytest.raises(ValueError, match=variable):
        oauth_db.build_url({**BASE, "PASSWORDEXCHANGE_DBDRIVER": driver, **overrides})


def test_make_engine_bounds_the_pool_and_pings():
    engine = oauth_db.make_engine({**BASE, "PASSWORDEXCHANGE_DBDRIVER": "postgres"})
    try:
        assert engine.pool.size() == 2
        assert engine.pool._max_overflow == 2
        assert engine.pool._recycle == 1800
        assert engine.pool._pre_ping is True
    finally:
        engine.dispose()


def test_make_engine_uses_the_url_from_build_url():
    env = {**BASE, "PASSWORDEXCHANGE_DBDRIVER": "postgres", "PASSWORDEXCHANGE_DBPORT": "6543"}
    engine = oauth_db.make_engine(env)
    try:
        assert engine.url == oauth_db.build_url(env)
        assert engine.dialect.name == "postgresql"
    finally:
        engine.dispose()


@settings(max_examples=200)
@given(user=text, password=text, db=text, host=hostname, p=port)
def test_url_object_carries_arbitrary_values_verbatim_for_both_drivers(user, password, db, host, p):
    for driver, dbhost in (("postgres", host), ("mysql", f"{host}:{p}")):
        env = {
            "PASSWORDEXCHANGE_DBDRIVER": driver,
            "OAUTHDB_USER": user,
            "OAUTHDB_PASSWORD": password,
            "OAUTHDB_NAME": db,
            "PASSWORDEXCHANGE_DBHOST": dbhost,
            "PASSWORDEXCHANGE_DBPORT": str(p),
        }
        url = oauth_db.build_url(env)
        assert (url.username, url.password, url.host, url.port, url.database) == (
            user, password, host, p, db,
        )
