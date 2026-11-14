output "instance_connection_name" {
  description = "Connection name for the cloud-sql-proxy sidecar args"
  value       = data.google_sql_database_instance.shared.connection_name
}

output "proxy_gsa_email" {
  description = "GSA for the iam.gke.io/gcp-service-account annotation on KSA password-exchange"
  value       = google_service_account.proxy.email
}

output "db_users" {
  description = "Database user => database name"
  value       = local.users
}

output "db_passwords" {
  description = "Database user => password. Read with: terraform output -json db_passwords"
  value       = { for u, p in random_password.db : u => p.result }
  sensitive   = true
}
