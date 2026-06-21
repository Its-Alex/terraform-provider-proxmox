data "proxmox_notification_endpoint_sendmail" "ops" {
  name = "ops-sendmail"
}

output "ops_recipients" {
  value = data.proxmox_notification_endpoint_sendmail.ops.mailto
}
