# password-exchange Cloud SQL Terraform

Databases, users and the Cloud SQL Auth Proxy identity for password-exchange on the **shared `sre-bible` Cloud SQL instance** (POSTGRES_18, `db-f1-micro`, ZONAL). Epic #562, issue #563.

**The instance is owned by sre-bible's Terraform** (`sre-bible/deploy/terraform`, sre-bible ADR 0015). This state only reads it through `data.google_sql_database_instance.shared`. Tier, flags, backups, PITR, maintenance window and deletion protection all live there. Never add a `google_sql_database_instance` resource here.

## What it creates

| Resource | Names |
|---|---|
| `google_sql_database.this` | `passwordexchange` (prod), `slack_oauth` (prod slackbot), `passwordexchange_dev` (dev) |
| `random_password.db` + `google_sql_user.this` | `passwordexchange`, `slackbot`, `passwordexchange_dev` (32 chars, no special characters) |
| `google_service_account.proxy` | `password-exchange-sql` |
| `google_project_iam_member.cloudsql_client` | `roles/cloudsql.client` for the GSA |
| `google_service_account_iam_member.workload_identity` | KSA `default/password-exchange` may act as the GSA |

There is one Slack app and its slackbot runs only in prod, so there is no dev OAuth database.

Databases have `prevent_destroy`. All `google_sql_user` users are `cloudsqlsuperuser` members and can connect to each other's databases (and to `sre_bible`); that was accepted, and there are no per-role connection limits. Pool caps in the Secrets are the connection budget (#568).

## Apply

```sh
cd deploy/terraform
terraform init
terraform test     # offline contract tests (mocked providers)
terraform plan     # expect 12 to add on a fresh state, never an instance change
terraform apply
```

State: `gs://gen-lang-client-0479899208-tfstate/password-exchange` (sre-bible uses the `sre-bible` prefix in the same bucket).

Read the outputs for the Secrets:

```sh
terraform output instance_connection_name
terraform output proxy_gsa_email
terraform output -json db_passwords
```

## Kubernetes Secrets

`password-exchange-db-dev` and `password-exchange-db-prod` in namespace `default` hold the DB settings per phase, so cutover and rollback are a Secret change plus a rollout restart. They're SOPS-encrypted with the key in `kubernetes/sops.yaml` and applied by hand:

```sh
sops -d kubernetes/password-exchange-db-dev.encrypted.yaml  | kubectl apply -f -
sops -d kubernetes/password-exchange-db-prod.encrypted.yaml | kubectl apply -f -
```

The KSA `password-exchange` needs the annotation `iam.gke.io/gcp-service-account: <proxy_gsa_email>` (added to `k8s/` in #568).

## Shared instance: things to know

- **Connections.** `db-f1-micro` has 25 `max_connections`, about 19 usable by apps. sre-bible takes 8 (2 × 4). password-exchange's pool caps are budgeted in #568. If `too many clients` shows up, raising the tier (in sre-bible's Terraform) is the escape hatch.
- **No HA.** The instance is ZONAL. A zone outage or restart takes both apps down.
- **Restarts.** Instance-level changes in sre-bible's Terraform can restart the instance and affect password-exchange too.
- **Restore is per instance.** To recover password-exchange alone: PITR-clone `sre-bible` to a new instance at the target time, `pg_dump` the password-exchange databases from the clone, restore them into `sre-bible`, then delete the clone. Never restore over the shared instance.
- **Teardown.** Postgres won't drop a user that owns objects. Drop or reassign the database first, and remove `prevent_destroy` deliberately.
