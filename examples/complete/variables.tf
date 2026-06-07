variable "nightingale_endpoint" {
  description = "Base URL for the Nightingale center API."
  type        = string
}

variable "nightingale_token" {
  description = "User token sent as X-User-Token."
  type        = string
  sensitive   = true
}

variable "busi_group_id" {
  description = "Business group ID for Nightingale resources."
  type        = number
  default     = 1
}

variable "datasource_id" {
  description = "Prometheus datasource ID."
  type        = number
  default     = 1
}
