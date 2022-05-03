
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

# slackclient = WebClient(token=os.environ.get("SLACK_BOT_TOKEN"))

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
    logger.error("Something went wrong creating intial dbs" + e)

bolt_app = App(
    logger=logger,
    signing_secret=os.environ.get("SLACK_SIGNING_SECRET"),
    installation_store=installation_store,
    oauth_settings=OAuthSettings(
        client_id=client_id,
        client_secret=client_secret,
        install_page_rendering_enabled=False,
        state_store=oauth_state_store,
        scopes=["chat:write", "commands", "groups:history", "groups:write", "im:history", "mpim:history", "channels:history", "groups:read"]
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
@bolt_app.message(re.compile("password:(?!\s+$)(?!\s*&lt;[rs].+&gt;)(?!\s*```).+"))
def reply_in_thread(client, ack, payload, body, logger, say, context):
    """ This will reply in thread instead of creating a new thread """
    ack()

    user = payload.get("user")
    channel = payload.get("channel")
    response = client.chat_postMessage( channel=channel,
                                     thread_ts=payload.get('ts'),
                                     text=f"Hi <@{user}>, you should use the `/password` command for sharing passwords.")

@bolt_app.command("/password")
@bolt_app.command("/encrypt")
def encrypt_command(payload: dict, ack, respond):
    ack()
    slack_text=payload.get('text')
    if slack_text =="help":
        respond(response_type="ephemeral", text="• Use `/encrypt` or `/password` to share sensitive information\n • If you want to share redacted passwords use `password: <redacted>` or `password: <snipped>` ")
        return
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

    respond(text=text, response_type="ephemeral")
@bolt_app.action("share_to_channel")
def post_to_channel(ack, payload, logger, respond):
    ack()
    respond(response_type="in_channel", delete_original="true", text=f"{payload['value']}" )
    
@bolt_app.event("app_home_opened")
def update_home_tab(client, event, logger):
    try:
        # Call views.publish with the built-in client
        client.views_publish(
            # Use the user ID associated with the event
            user_id=event["user"],
            # Home tabs must be enabled in your app configuration
            view={
                "type": "home",
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": f"Hi  <@{event['user']}>, :wave:"
			}
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "Great to see you here! Password exchange lets you share passwords and sensitive information securely in slack"
			}
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "• Use `/password`  or `/encrypt` to share passwords \n • Slackbot will remind users if it detects a password shared unencrypted\n • Use `password: <redacted>` to show passwords that were removed \n"
			}
		}
	]
}
        )
    except Exception as e:
        logger.error(f"Error publishing home tab: {e}")
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
