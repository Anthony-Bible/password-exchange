"""Create the slack-sdk tables in OAUTHDB_NAME exactly as slackbot/program.py does (#569).

Run before the load: prod has no tables yet and pgloader runs data-only.
Env: SLACKBOT_DIR plus the slackbot DB settings (PASSWORDEXCHANGE_DB*, OAUTHDB_*).
"""
import os, sys
sys.path.insert(0, os.environ["SLACKBOT_DIR"])
import oauth_db
from slack_sdk.oauth.installation_store.sqlalchemy import SQLAlchemyInstallationStore
from slack_sdk.oauth.state_store.sqlalchemy import SQLAlchemyOAuthStateStore
engine = oauth_db.make_engine(os.environ)
SQLAlchemyInstallationStore(client_id="unused", engine=engine).metadata.create_all(engine)
SQLAlchemyOAuthStateStore(expiration_seconds=120, engine=engine).metadata.create_all(engine)
print("created on", engine.url.render_as_string(hide_password=True))
