# Contract tests for the password-exchange Cloud SQL Terraform (issue #563).
# Providers are mocked, so this runs offline: `terraform test` from deploy/terraform.

mock_provider "google" {
  override_data {
    target = data.google_sql_database_instance.shared
    values = {
      name            = "sre-bible"
      connection_name = "gen-lang-client-0479899208:us-central1:sre-bible"
    }
  }
}

mock_provider "random" {}

run "databases_live_on_the_shared_instance" {
  command = plan

  assert {
    condition = toset([for db in google_sql_database.this : db.name]) == toset([
      "passwordexchange", "slack_oauth", "passwordexchange_dev",
    ])
    error_message = "expected exactly the prod app, prod Slack OAuth and dev app databases"
  }

  assert {
    condition     = alltrue([for db in google_sql_database.this : db.instance == "sre-bible"])
    error_message = "every database must be created on the shared sre-bible instance"
  }
}

run "one_user_per_workload" {
  command = plan

  assert {
    condition = toset([for u in google_sql_user.this : u.name]) == toset([
      "passwordexchange", "slackbot", "passwordexchange_dev",
    ])
    error_message = "expected users passwordexchange, slackbot and passwordexchange_dev"
  }

  assert {
    condition     = alltrue([for u in google_sql_user.this : u.instance == "sre-bible"])
    error_message = "every user must be created on the shared sre-bible instance"
  }

  assert {
    condition     = alltrue([for p in random_password.db : p.length >= 32 && p.special == false])
    error_message = "passwords must be at least 32 characters and URL-safe (no special characters)"
  }
}

run "proxy_identity_uses_workload_identity" {
  command = plan

  assert {
    condition     = google_service_account.proxy.account_id == "password-exchange-sql"
    error_message = "proxy GSA must be password-exchange-sql, separate from sre-bible-proxy"
  }

  assert {
    condition     = google_project_iam_member.cloudsql_client.role == "roles/cloudsql.client"
    error_message = "proxy GSA needs roles/cloudsql.client"
  }

  assert {
    condition     = google_service_account_iam_member.workload_identity.member == "serviceAccount:gen-lang-client-0479899208.svc.id.goog[default/password-exchange]"
    error_message = "Workload Identity must bind KSA default/password-exchange"
  }
}

run "outputs_feed_the_k8s_secrets" {
  command = plan

  assert {
    condition     = output.instance_connection_name == "gen-lang-client-0479899208:us-central1:sre-bible"
    error_message = "instance_connection_name must come from the shared instance data source"
  }

  assert {
    condition     = keys(output.db_users) == ["passwordexchange", "passwordexchange_dev", "slackbot"]
    error_message = "db_users output must list every user"
  }
}
