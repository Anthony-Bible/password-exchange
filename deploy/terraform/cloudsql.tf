# The instance is owned by sre-bible's Terraform (sre-bible ADR 0015). Read it
# only: this state must never be able to modify or destroy it. Never add a
# google_sql_database_instance resource here.
data "google_sql_database_instance" "shared" {
  name = var.instance_name
}

locals {
  # One user per workload, each with its own database. The slackbot exists
  # only in prod, so there is no dev OAuth database.
  users = {
    passwordexchange     = "passwordexchange"     # prod database + reminder
    slackbot             = "slack_oauth"          # prod slackbot OAuth stores
    passwordexchange_dev = "passwordexchange_dev" # dev database + reminder
  }
}

resource "google_sql_database" "this" {
  for_each = toset(values(local.users))

  instance = data.google_sql_database_instance.shared.name
  name     = each.key

  # Holds live secrets and Slack installations. Remove this deliberately
  # (and back up first) to tear down.
  lifecycle {
    prevent_destroy = true
  }
}

resource "random_password" "db" {
  for_each = local.users

  length  = 32
  special = false # goes into DSNs and SQLAlchemy URLs unescaped
}

resource "google_sql_user" "this" {
  for_each = local.users

  instance = data.google_sql_database_instance.shared.name
  name     = each.key
  password = random_password.db[each.key].result
}
