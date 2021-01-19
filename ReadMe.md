# Password.exchange
Password exchange was built because there was no known way to securely share passwords. No need to have someone pick up the phone or set up "complicated" gpg. While password exchange focuses on passwords it in theory can be used for any text. 
---
## How it works
You fill out the form with the neccessary information including yours and their names and emails. Once you click send, we send two emails, one to the recipeint with the information to retrieve their password and one to you to help track when the recipient opens the email and visits the page. 
---
### Features
1. Send message to both users. 
2. Remind users after a day of not opening
3. Expire after 7 days or 1 hour after viewing
   a. In the future this will be configurable


TODO:
  1. Is Client Side encryption feasable?
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