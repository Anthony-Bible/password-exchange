
import logging

logging.basicConfig(level=logging.INFO)
logging.getLogger("sqlalchemy.engine").setLevel(logging.INFO)

import base64
import os
import re
import sys
from flask import Flask, request
from slack_sdk import WebClient
from slack_bolt import App, Say
from slack_bolt.adapter.flask import SlackRequestHandler
#USE SQLAlchemy for oauth store
from slack_bolt.oauth.oauth_settings import OAuthSettings
from slack_sdk.oauth.installation_store.sqlalchemy import SQLAlchemyInstallationStore
from slack_sdk.oauth.state_store.sqlalchemy import SQLAlchemyOAuthStateStore

import sqlalchemy
from sqlalchemy.engine import Engine
import MySQLdb
# gazelle:ignore encryptionClient
import encryptionClient

logger = logging.getLogger(__name__)

slackclient = WebClient(token=os.environ.get("SLACK_BOT_TOKEN"))

oauthpassword = os.environ.get("OAUTHDB_PASSWORD")
oauthUser = os.environ.get("OAUTHDB_USER")
oauthdb = os.environ.get("OAUTHDB_NAME")
dbhost = os.environ.get("PASSWORDEXCHANGE_DBHOST")
client_id = os.environ.get("SLACK_CLIENT_ID")
client_secret = os.environ.get("SLACK_CLIENT_SECRET")
database_url = "mysql+mysqldb://" + oauthUser +":" + oauthpassword +"@" + dbhost +"/" + oauthdb

engine: Engine = sqlalchemy.create_engine(database_url, pool_pre_ping=True)

#create oauth and installation store
installation_store = SQLAlchemyInstallationStore(
    client_id=client_id,
    engine=engine,
    logger=logger,
)
oauth_state_store = SQLAlchemyOAuthStateStore(
    expiration_seconds=120,
    engine=engine,
    logger=logger,
)

try:
    installation_store.metadata.create_all(engine)
    oauth_state_store.metadata.create_all(engine)
except Exception as e:
    print("Something went wrong creating intial dbs" + e)

bolt_app = App(
    logger=logger,
    signing_secret=os.environ.get("SLACK_SIGNING_SECRET"),
    installation_store=installation_store,
    oauth_settings=OAuthSettings(
        client_id=client_id,
        client_secret=client_secret,
        state_store=oauth_state_store,
        scopes=["chat:write", "commands", "groups:history", "im:history", "mpim:history", "channels:history", "groups:read"]
    ),
)
app = Flask(__name__)
handler = SlackRequestHandler(bolt_app)

# Match all messages that "password:" but not ones with just whitespace
# @bolt_app.message(re.compile("password:(?!\s+$)(?!\s*<[sr].+>).+|PASSWORD:(?!\s+$)(?!\s*<[sr].+>).+"))
# @bolt_app.message("hello slacky")
# def greetings(ack, payload, body, logger, say):
#     """ This will check all the message and pass only those which has 'hello slacky' in it """
#     ack()
#     user = payload.get("user")
#     print("this is the payload")
#     logger.info(payload)
#     logger.info(body)
#     say(f"Hi <@{user}>, you should use the `/password` command for sharing passwords")

client = encryptionClient.EncryptionServiceClient()
@bolt_app.message(re.compile("password:(?!\s+$)(?!\s*&lt;[rs].+&gt;)(?!\s*```).+",re.DEBUG))
def reply_in_thread(ack, payload, body, logger, say, context):
    """ This will reply in thread instead of creating a new thread """
    ack()
    user = payload.get("user")
    response = slackclient.chat_postMessage(channel=payload.get('channel'),
                                     thread_ts=payload.get('ts'),
                                     text=f"Hi <@{user}>, you should use the `/password` command for sharing passwords.")
    
@bolt_app.command("/password")
@bolt_app.command("/encrypt")
def encrypt_command(payload: dict, ack, respond):
    ack()
    slack_text=payload.get('text')
    key, guid = client.encrypt_text(slack_text)
    #TODO: put encoding to base64 in a separate function
    #slteHost + "decrypt/" + guid.String() + "/" + string(b64.URLEncoding.EncodeToString(encryptionRequest.Key)),
    # message_bytes = key.encode('ascii')
    base64_bytes = base64.urlsafe_b64encode(key)
    base64_key = base64_bytes.decode('ascii')
    sitehost = os.environ['PASSWORDEXCHANGE_HOST']
    decrypt_url =   (sitehost + "decrypt/" + guid + "/" + base64_key)
    text = {
        "blocks": [
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"<@{payload['user_name']}> here's the encrypted url: " + decrypt_url,
                }
            },
            {
                "type": "actions",
                "elements": [
                    {
                        "type": "button",
                        "action_id": "share_to_channel",
                        "accessibility_label": "Do you want to share this to the channel?",
                        "text": {
                            "type": "plain_text",
                            "text": "Share to channel"
                        },
                        "style": "danger",
                        "value": decrypt_url
                    }
                ]
            }
        ]
    }

    respond(text=text)
@bolt_app.action("share_to_channel")
def post_to_channel(ack, payload, logger, respond):
    ack()
    logger.info(payload)
    respond(response_type="in_channel", delete_original="true", text=f"{payload['value']}" )
@app.route("/slack/events", methods=["POST"])
def slack_events():
    """ Declaring the route where slack will post a request """
    return handler.handle(request)

@app.route("/slack/install", methods=["GET"])
def install():
    return handler.handle(request)

@app.route("/slack/oauth_redirect", methods=["GET"])
def oauth_redirect():
    return handler.handle(request)


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=3000, debug=False)
