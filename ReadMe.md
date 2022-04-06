# Password.exchange
Password exchange was built because there was no known way to securely share passwords. No need to have someone pick up the phone or set up "complicated" gpg. While password exchange focuses on passwords it in theory can be used for any text. 

---

## How it works
### Website

You fill out the form with the neccessary information including both of your names and emails. We just use the email to send the link to the  recipeint, in the near future we will make the emails optional. 

### slackbot
 To install our slackbot go to (https://api.password.exchange/slack/install). If you have setup your own version of this app,  you can go to https://yoursite.com/slack/install. 

 Once installed to your organization, you can use the `/encrypt` command which will send the text to the bot and the bot will send a link to access the unencrypted text. 

 **NOTE:** Slackbot relies on the website since the encryption and database service are deployed with it. 

---

### Features
1. [planned] Send message to both users. 
2. [planned] Remind users after a day of not opening
3. [planned] Get notifications of the following
   1. email opened, page visited, Page viewed
3. [planned] Expire after 7 days or 1 hour after viewing
   1. In the future this will be configurable


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
#### Kubernetes
1. Run `bazel build //...`
2. To deploy kubernetes manifests `bazel run //kubernetes:deployments.create`
3. To Reapply a kubernetes manifest (after a code change) `bazel run //kubernetes:deployments.apply`
 maybe skaffold?
##### Docker
1. 



##### Kubernetes
1. Edit the deployment.yaml for any needed changes
2. Apply the deployment.yaml
```bash 
kubectl apply -f kubernetes/
```
