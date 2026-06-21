data "proxmox_notification_matcher" "backup_failures" {
  name = "backup-failures"
}

output "backup_failure_targets" {
  value = data.proxmox_notification_matcher.backup_failures.target
}
