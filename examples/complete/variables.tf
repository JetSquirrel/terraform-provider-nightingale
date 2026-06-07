variable "nightingale_endpoint" {
  description = "Nightingale API endpoint URL."
  type        = string
}

variable "nightingale_token" {
  description = "Nightingale API token."
  type        = string
  sensitive   = true
}

variable "busi_group_id" {
  description = "Business group ID."
  type        = number
  default     = 1
}

variable "datasource_id" {
  description = "Prometheus datasource ID."
  type        = number
  default     = 1
}
