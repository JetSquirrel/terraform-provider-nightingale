variable "nightingale_endpoint" {
  description = "Nightingale API endpoint URL."
  type        = string
}

variable "nightingale_token" {
  description = "Nightingale API token."
  type        = string
  sensitive   = true
}
