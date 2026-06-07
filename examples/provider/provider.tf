provider "nightingale" {
  endpoint = var.nightingale_endpoint
  token    = var.nightingale_token

  # Optional settings
  # timeout_seconds          = 30
  # insecure_skip_tls_verify = false
}
