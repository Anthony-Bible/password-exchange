import base64
import os
import re
import sys
from flask import Flask, request
from slack_sdk import WebClient
from slack_bolt import App, Say

from slack_bolt.adapter.flask import SlackRequestHandler
# gazelle:ignore encryptionClient
import encryptionClient
app = Flask(__name__)
slackclient = WebClient(token=os.environ.get("SLACK_BOT_TOKEN"))
bolt_app = App(token=os.environ.get("SLACK_BOT_TOKEN"), signing_secret=os.environ.get("SLACK_SIGNING_SECRET"))

@bolt_app.message("hello slacky")
def greetings(payload: dict, say: Say):
    """ This will check all the message and pass only those which has 'hello slacky' in it """
    user = payload.get("user")
    say(f"Hi <@{user}>")

client = encryptionClient.EncryptionServiceClient()

@bolt_app.message(re.compile("(hi|hello|hey) slacky"))
def reply_in_thread(payload: dict):
    """ This will reply in thread instead of creating a new thread """
    response = slackclient.chat_postMessage(channel=payload.get('channel'),
                                     thread_ts=payload.get('ts'),
                                     text=f"Hi<@{payload['user']}>")
@bolt_app.command("/encrypt")
def encrypt_command(say, payload: dict, ack):
    ack()
    print(payload)
    slack_text=payload.get('text')
    sys.stdout.flush()                      # <--- added line to flush output

    key, guid = client.encrypt_text(slack_text)
    #TODO: put encoding to base64 in a separate function
    #slteHost + "decrypt/" + guid.String() + "/" + string(b64.URLEncoding.EncodeToString(encryptionRequest.Key)),
    # message_bytes = key.encode('ascii')
    base64_bytes = base64.urlsafe_b64encode(key)
    base64_key = base64_bytes.decode('ascii')
    sitehost = os.environ['PASSWORDEXCHANGE_HOST']
    text = {
        "blocks": [
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"<@{payload['user_name']}> here's the encrypted url:" + (sitehost + "decrypt/" + guid + "/" + base64_key)
                }
            }
        ]
    }

    say(text=text)

@app.route("/slack/events", methods=["POST"])
def slack_events():
    """ Declaring the route where slack will post a request """
    return handler.handle(request)

handler = SlackRequestHandler(bolt_app)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=3000, debug=False)
