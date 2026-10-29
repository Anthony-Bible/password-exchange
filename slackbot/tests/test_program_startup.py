"""Startup behaviour of the real slackbot/program.py, run as a subprocess the
way the Dockerfile does (`python3 slackbot/program.py`)."""
import json
import logging
import subprocess
import sys

from conftest import SLACKBOT_DIR, TESTS_DIR

UNREACHABLE = {
    "PASSWORDEXCHANGE_DBDRIVER": "postgres",
    "OAUTHDB_USER": "u",
    "OAUTHDB_PASSWORD": "p",
    "OAUTHDB_NAME": "d",
    "PASSWORDEXCHANGE_DBHOST": "127.0.0.1",
    "PASSWORDEXCHANGE_DBPORT": "1",
}


def run_program(env, timeout=60):
    return subprocess.run(
        [sys.executable, str(SLACKBOT_DIR / "program.py")],
        env=env, capture_output=True, text=True, timeout=timeout,
    )


def test_create_all_failure_aborts_startup_with_traceback(program_env):
    proc = run_program({**program_env, **UNREACHABLE})
    assert proc.returncode != 0
    assert "Something went wrong creating initial dbs" in proc.stderr
    assert "Traceback" in proc.stderr
    # The original DB error must survive, not be replaced by a TypeError from
    # string + exception concatenation.
    assert "TypeError" not in proc.stderr


def test_bogus_driver_aborts_startup_naming_the_variable(program_env):
    proc = run_program({**program_env, **UNREACHABLE, "PASSWORDEXCHANGE_DBDRIVER": "bogus"})
    assert proc.returncode != 0
    assert "PASSWORDEXCHANGE_DBDRIVER" in proc.stderr


def test_find_bot_logs_no_parameter_values(program_env, postgres_env):
    proc = subprocess.run(
        [sys.executable, str(TESTS_DIR / "_program_probe.py")],
        env={**program_env, **postgres_env}, capture_output=True, text=True, timeout=120,
    )
    assert proc.returncode == 0, proc.stderr
    out = json.loads(proc.stdout)
    assert out["found_token"] == "xoxb-SUPER-SECRET-TOKEN"
    assert out["engine_logger_level"] >= logging.WARNING
    blob = json.dumps(out["sqlalchemy_records"])
    assert "xoxb-SUPER-SECRET-TOKEN" not in blob
    assert "T-PROBE" not in blob
    assert "xoxb-SUPER-SECRET-TOKEN" not in proc.stderr
