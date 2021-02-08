# Password.exchange
Password exchange was built because there was no known way to securely share passwords. No need to have someone pick up the phone or set up "complicated" gpg. While password exchange focuses on passwords it in theory can be used for any text. 

---

## How it works
You fill out the form with the neccessary information including yours and their names and emails. Once you click send, we send two emails, one to the recipeint with the information to retrieve their password and one to you to help track when the recipient opens the email and visits the page. 

---

### Features
1. [planned] Send message to both users. 
2. [planned] Remind users after a day of not opening
3. [planned] Get notifications of the following
   a. email opened, page visited, Page viewed
3. [planned] Expire after 7 days or 1 hour after viewing
   a. In the future this will be configurable


TODO:
  1. Allow user to generate password 
  2. Is Client Side encryption feasable?

Future (hopeful) Intergrations:
1. Bitwarden
2. Google drive (files)
3. Salesforce
4. Lastpass
5. Email (pgp)
  a. User can send pgp encrypted email, we retrieve, decrypt and then send like the regular process
6. slack(?)

---

### Installation
##### Docker
1. Clone this repo locally
2. run docker compose
```bash
 docker-compose up -d
```
##### Kubernetes
1. Edit the deployment.yaml for any needed changes
2. Apply the deployment.yaml
```bash 
kubectl apply -f kubernetes/
```