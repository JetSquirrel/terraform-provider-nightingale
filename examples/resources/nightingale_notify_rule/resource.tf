resource "nightingale_notify_rule" "email_ops" {
  name           = "Email OPS"
  enable         = true
  user_group_ids = [1]

  notify_configs = [{
    channel_id  = 1
    template_id = 1
  }]
}

resource "nightingale_notify_rule" "webhook_devops" {
  name           = "Webhook DevOps"
  enable         = true
  user_group_ids = [2, 3]

  notify_configs = [{
    channel_id  = 2
    template_id = 2
    params = {
      webhook_url = "https://hooks.example.com/alerts"
      secret      = "my-secret-token"
    }
  }]
}
