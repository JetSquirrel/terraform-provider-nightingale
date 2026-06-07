resource "nightingale_notify_rule" "example" {
  name           = "Email OPS Team"
  enable         = true
  user_group_ids = [1]

  notify_configs = [{
    channel_id  = 1
    template_id = 1
  }]
}
