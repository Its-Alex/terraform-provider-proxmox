data "proxmox_notification_targets" "all" {}

output "all_target_names" {
  value = [for t in data.proxmox_notification_targets.all.targets : t.name]
}

output "user_created_sendmail_targets" {
  value = [
    for t in data.proxmox_notification_targets.all.targets :
    t.name if t.type == "sendmail" && t.origin == "user-created"
  ]
}
