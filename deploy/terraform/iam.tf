# Identity for the Cloud SQL Auth Proxy sidecars. Separate from sre-bible-proxy
# so each app's access can be revoked on its own.
resource "google_service_account" "proxy" {
  account_id   = "password-exchange-sql"
  display_name = "password-exchange Cloud SQL Auth Proxy"
}

# Project-wide: this also allows dialling other instances in the project.
# Database users and passwords are what actually scope access.
resource "google_project_iam_member" "cloudsql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.proxy.email}"
}

resource "google_service_account_iam_member" "workload_identity" {
  service_account_id = google_service_account.proxy.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project_id}.svc.id.goog[${var.k8s_namespace}/${var.k8s_service_account}]"
}
