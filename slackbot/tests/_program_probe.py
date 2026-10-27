"""Subprocess helper: import the real program.py, save + look up a bot, and
report every log record the sqlalchemy.* loggers actually emitted."""
import json
import logging
import sys

import program  # noqa: F401  (module-scope wiring is what is under test)
from slack_sdk.oauth.installation_store import Installation

records = []


class Collect(logging.Handler):
    def emit(self, record):
        records.append({"logger": record.name, "msg": record.getMessage()})


handler = Collect(level=logging.DEBUG)
logging.getLogger().addHandler(handler)

store = program.installation_store
store.save(
    Installation(
        app_id="A1", enterprise_id=None, team_id="T-PROBE", user_id="U1",
        bot_token="xoxb-SUPER-SECRET-TOKEN", bot_id="B1", bot_user_id="UB1",
        bot_scopes=["chat:write"], installed_at=1_700_000_000.0,
    )
)
bot = store.find_bot(enterprise_id=None, team_id="T-PROBE")
sa = [r for r in records if r["logger"].startswith("sqlalchemy")]
json.dump(
    {
        "found_token": bot.bot_token if bot else None,
        "engine_logger_level": logging.getLogger("sqlalchemy.engine").getEffectiveLevel(),
        "sqlalchemy_records": sa,
    },
    sys.stdout,
)
