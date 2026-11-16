"""find_bot/find_installation for every team through the slackbot's own store (#569).

Prints only found/missing per team, never ids or tokens. Exit 1 if any lookup fails.
Env: SLACKBOT_DIR, SLACK_CLIENT_ID and the slackbot DB settings.
"""
import os, sys
sys.path.insert(0, os.environ["SLACKBOT_DIR"])
import oauth_db
from sqlalchemy import text
from slack_sdk.oauth.installation_store.sqlalchemy import SQLAlchemyInstallationStore
engine = oauth_db.make_engine(os.environ)
store = SQLAlchemyInstallationStore(client_id=os.environ["SLACK_CLIENT_ID"], engine=engine)
with engine.connect() as c:
    teams = c.execute(text("select distinct enterprise_id, team_id, is_enterprise_install from slack_installations where client_id=:c"), {"c": os.environ["SLACK_CLIENT_ID"]}).all()
ok = True
for i, (ent, team, ent_inst) in enumerate(teams, 1):
    bot = store.find_bot(enterprise_id=ent, team_id=team, is_enterprise_install=ent_inst)
    inst = store.find_installation(enterprise_id=ent, team_id=team, is_enterprise_install=ent_inst)
    good = bool(bot and bot.bot_token and inst and inst.bot_token)
    ok &= good
    print(f"team {i}: bot={'found' if bot else 'MISSING'} token={'yes' if bot and bot.bot_token else 'no'} installation={'found' if inst else 'MISSING'}")
print("SLACK LOOKUPS:", "PASS" if ok and teams else "FAIL", f"({len(teams)} teams)")
sys.exit(0 if ok and teams else 1)
