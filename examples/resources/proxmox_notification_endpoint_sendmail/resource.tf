resource "proxmox_notification_endpoint_sendmail" "ops" {
  name         = "ops-sendmail"
  mailto       = ["ops@example.com", "oncall@example.com"]
  mailto_user  = ["root@pam"]
  from_address = "alerts@example.com"
  author       = "Proxmox VE"
  comment      = "Primary on-call alerting"
}
