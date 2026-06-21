resource "proxmox_notification_endpoint_sendmail" "ops" {
  name   = "ops-sendmail"
  mailto = ["ops@example.com"]
}

resource "proxmox_notification_matcher" "backup_failures" {
  name           = "backup-failures"
  target         = [proxmox_notification_endpoint_sendmail.ops.name]
  match_field    = ["exact:type=vzdump"]
  match_severity = ["error", "warning"]
  match_calendar = ["mon..fri 09:00-17:00"]
  mode           = "all"
  comment        = "Route backup failures to ops during business hours"
}
