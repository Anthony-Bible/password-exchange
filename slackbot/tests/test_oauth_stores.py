"""slack-sdk SQLAlchemy stores behave identically on Postgres 17 and MySQL 8.0
when the engine comes from oauth_db.make_engine."""
import logging
import time

import pytest
import sqlalchemy
from slack_sdk.oauth.installation_store import Installation
from slack_sdk.oauth.installation_store.sqlalchemy import SQLAlchemyInstallationStore
from slack_sdk.oauth.state_store.sqlalchemy import SQLAlchemyOAuthStateStore

import oauth_db

CLIENT_ID = "111.222"
log = logging.getLogger("test")


@pytest.fixture(params=["postgres", "mysql"])
def engine(request):
    env = request.getfixturevalue(f"{request.param}_env")
    eng = oauth_db.make_engine(env)
    yield eng
    meta = sqlalchemy.MetaData()
    meta.reflect(bind=eng)
    meta.drop_all(bind=eng)
    eng.dispose()


def make_stores(engine, expiration_seconds=120):
    inst = SQLAlchemyInstallationStore(client_id=CLIENT_ID, engine=engine, logger=log)
    state = SQLAlchemyOAuthStateStore(expiration_seconds=expiration_seconds, engine=engine, logger=log)
    return inst, state


def installation(token="xoxb-first", team="T123", **kw):
    return Installation(
        app_id="A1",
        enterprise_id=None,
        team_id=team,
        user_id="U1",
        bot_token=token,
        bot_id="B1",
        bot_user_id="UB1",
        bot_scopes=["chat:write", "commands"],
        is_enterprise_install=False,
        installed_at=1_700_000_000.0,
        **kw,
    )


def test_create_all_is_idempotent(engine):
    inst, state = make_stores(engine)
    for _ in range(2):
        inst.metadata.create_all(engine)
        state.metadata.create_all(engine)
    names = set(sqlalchemy.inspect(engine).get_table_names())
    assert {"slack_installations", "slack_bots", "slack_oauth_states"} <= names


def test_installation_roundtrips_fields(engine):
    inst, state = make_stores(engine)
    inst.metadata.create_all(engine)
    saved = installation()
    inst.save(saved)

    found = inst.find_installation(enterprise_id=None, team_id="T123")
    assert found is not None
    assert found.team_id == saved.team_id
    assert found.user_id == saved.user_id
    assert found.bot_token == saved.bot_token
    assert found.is_enterprise_install is False
    assert found.installed_at == pytest.approx(saved.installed_at, abs=1)

    bot = inst.find_bot(enterprise_id=None, team_id="T123")
    assert bot is not None
    assert bot.bot_token == saved.bot_token
    assert bot.is_enterprise_install is False
    assert bot.installed_at == pytest.approx(saved.installed_at, abs=1)


def test_second_save_for_same_team_returns_latest(engine):
    inst, _ = make_stores(engine)
    inst.metadata.create_all(engine)
    inst.save(installation(token="xoxb-old"))
    inst.save(installation(token="xoxb-new"))

    assert inst.find_installation(enterprise_id=None, team_id="T123").bot_token == "xoxb-new"
    assert inst.find_bot(enterprise_id=None, team_id="T123").bot_token == "xoxb-new"


def test_unknown_team_is_not_found(engine):
    inst, _ = make_stores(engine)
    inst.metadata.create_all(engine)
    assert inst.find_bot(enterprise_id=None, team_id="TNOPE") is None


def test_state_is_consumable_exactly_once(engine):
    _, state = make_stores(engine)
    state.metadata.create_all(engine)
    s = state.issue()
    assert state.consume(s) is True
    assert state.consume(s) is False


def test_unknown_state_is_rejected(engine):
    _, state = make_stores(engine)
    state.metadata.create_all(engine)
    assert state.consume("never-issued") is False


def test_expired_state_is_rejected(engine):
    _, state = make_stores(engine, expiration_seconds=1)
    state.metadata.create_all(engine)
    s = state.issue()
    time.sleep(2.2)
    assert state.consume(s) is False
