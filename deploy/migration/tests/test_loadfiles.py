"""Contract tests for pgloader load-file generation and sequence targets (#569)."""
import os
import stat
import sys
from urllib.parse import unquote, urlsplit

from hypothesis import given, strategies as st

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import loadfiles  # noqa: E402
import sequences  # noqa: E402

passwords = st.text(min_size=1, max_size=24).filter(lambda s: "\x00" not in s)


def creds(app_pw="a", oauth_pw="b", pg_app="c", pg_slack="d"):
    return {
        "MY_APP_USER": "passwordexchange", "MY_APP_PASS": app_pw, "MY_APP_DB": "passwordExchange",
        "MY_OAUTH_USER": "oauthUser", "MY_OAUTH_PASS": oauth_pw, "MY_OAUTH_DB": "oauth",
        "PG_APP_PASS": pg_app, "PG_SLACK_PASS": pg_slack,
    }


def uris(body):
    first = body.split("\n", 2)
    return first[0].split("FROM ", 1)[1].strip(), first[1].split("INTO ", 1)[1].strip()


@given(passwords, passwords)
def test_any_password_round_trips_through_the_uri(my_pw, pg_pw):
    files = loadfiles.render(creds(app_pw=my_pw, pg_app=pg_pw), suffix="")
    src, dst = uris(files["1-messages.load"])
    assert unquote(urlsplit(src).password) == my_pw
    assert unquote(urlsplit(dst).password) == pg_pw


@given(st.sampled_from(["", "_rehearsal"]))
def test_targets_follow_the_suffix_and_sources_never_change(suffix):
    files = loadfiles.render(creds(), suffix=suffix)
    for name, body in files.items():
        src, dst = uris(body)
        assert urlsplit(src).path in ("/passwordExchange", "/oauth")
        assert urlsplit(dst).path in (f"/passwordexchange{suffix}", f"/slack_oauth{suffix}")


def test_every_pass_is_data_only_keeps_target_constraints_and_lands_in_public():
    for body in loadfiles.render(creds(), suffix="").values():
        assert "data only" in body and "no foreign keys" in body
        assert "RENAME TO 'public'" in body


def test_load_order_and_table_selection():
    files = loadfiles.render(creds(), suffix="")
    assert list(files) == ["1-messages.load", "2-app-rest.load", "3-slack.load"]
    loaded = " ".join(files.values())
    for table in ["messages", "email_reminders", "file_upload_sessions", "slack_installations", "slack_bots"]:
        assert f"'{table}'" in loaded
    for never in ["schema_migrations", "slack_oauth_states"]:
        assert never not in loaded


def test_written_files_are_owner_only(tmp_path):
    loadfiles.write(loadfiles.render(creds(), suffix=""), tmp_path)
    for f in tmp_path.iterdir():
        assert stat.S_IMODE(f.stat().st_mode) == 0o600


@given(st.integers(min_value=1, max_value=10**9), st.one_of(st.none(), st.integers(0, 10**9)))
def test_setval_continues_after_both_auto_increment_and_max(auto_increment, src_max):
    value, is_called = sequences.setval_args(auto_increment, src_max)
    next_id = value + 1 if is_called else value
    assert next_id == max(auto_increment, (src_max or 0) + 1)
    assert value >= 1
