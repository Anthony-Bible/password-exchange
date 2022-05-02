|  [Documentation](https://github.com/Anthony-Bible/password-exchange/wiki) | [Building](#building-from-source) | [Running](#running)

# Password.exchange
Password exchange was built because there was no known way to securely share passwords. No need to have someone pick up the phone or set up "complicated" gpg. While password exchange was originally created for passwords it can be used for any text. 

**BE sure to visit our wiki for detailed information**

---

## How it works
### [Website](https://password.exchange)

You fill out [the form](https://password.exchange) with the neccessary information including both of your names and emails (optional). We use email to send the link to the content, but there is an option to disable emails. For your name(s), this is used to personalize and let the recipient know who sent them the link. There is no verification on names so you can use whatever to remain anonymous. 

**NOTE:** All messages expire after 7 days. This means you won't be able to view your message after 7 days and will have to resend it. 

### slackbot
 To install our slackbot go to (https://api.password.exchange/slack/install). If you have setup your own version of this app,  you can go to https://yoursite.com/slack/install. 

 Once installed to your organization, you can use the `/encrypt` command which will send the text to the bot and the bot will send a link to access the unencrypted text. 

 **NOTE:** Slackbot relies on the database and encryption services and deployments. You can remove the website deployment/service from the yaml if you only intend to deploy the slackbot.
---

### Features
1. [planned] Send message to both users. 
2. [planned] Remind users after a day of not opening
3. [planned] Get notifications of the following
   1. email opened, page visited
4. Expire after 7 days

   1.  [planned] 1 hour after viewing
   1.  [planned] In the future this will be configurable


TODO:
  1. Allow user to generate password 
  2. Is Client Side encryption feasable?
     1. yes
     2. We can use [this](https://web.archive.org/web/20220205052255/https://bitwarden.com/help/send-encryption/) as inspiration
        1. Basically we send the data already encrypted to the server to store
        2. This prohibits Slack and bot integrations from using Client side encryption

Future (hopeful) Intergrations:
1. Bitwarden
2. Google drive (files)
3. Salesforce
4. Lastpass
5. Email (pgp)
   1. User can send pgp encrypted email, we retrieve, decrypt and then send like the regular process

---


### BUILDING from source
***NOTE:***  *The build isn't completley hermetic yet, While I'm working on making it hermetic these packages need to be installed on the host: `zstd libssl-dev build-essential curl wget gcc mariadb-client libmariadb-dev clang`*

#### Kubernetes (bazel)

1. Run `bazel build //...`
2. If you want to just generate the yaml, run: `bazel run //k8s:deployment-and-services`
3. To deploy kubernetes manifests `bazel run //k8s:deployments-services.create`
4. To Reapply a kubernetes manifest (after a code change) `bazel run //k8s:deployments-and-services.apply`



---
### Running
*Currently we only support kubernetes. If you don't have a kubernetes cluster, you have two options. If you use docker-desktop you can [enable a local kubernetes](https://docs.docker.com/desktop/kubernetes/), otherwise look into setting up [minikube](https://minikube.sigs.k8s.io/docs/start/) which allows you to set up kubernetes on your local machine.*
1. Download the Mysql file from the root of the project
   A. Update passsword in create user statements
   B. Import the mysgql schema `mysql -u user -p < passwordexchange.sql`
2. edit `kubernetes/secrets.yaml` with your information
   
   1. [view here for avaible options](https://github.com/Anthony-Bible/password-exchange/wiki/Environment-Variables)
3. Download the latest manifest in releases
4. Do a `kubectl apply -f password-exchange.yaml`
   1. You can find this in the latest release. 
