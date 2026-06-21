//go:build acceptance || all

//testacc:tier=light
//testacc:resource=notifications

/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/test"
)

func TestAccResourceNotificationSendmailEndpoint(t *testing.T) {
	te := test.InitEnvironment(t)

	tests := []struct {
		name  string
		steps []resource.TestStep
	}{
		{"create with mailto, update, drop optional fields", []resource.TestStep{
			{
				Config: te.RenderConfig(`
					resource "proxmox_notification_endpoint_sendmail" "acc_basic" {
						name         = "acc-sendmail-basic"
						mailto       = ["ops@example.com"]
						from_address = "alerts@example.com"
						author       = "Proxmox VE"
						comment      = "initial"
					}`),
				Check: resource.ComposeTestCheckFunc(
					test.ResourceAttributes("proxmox_notification_endpoint_sendmail.acc_basic", map[string]string{
						"id":           "acc-sendmail-basic",
						"name":         "acc-sendmail-basic",
						"mailto.#":     "1",
						"mailto.0":     "ops@example.com",
						"from_address": "alerts@example.com",
						"author":       "Proxmox VE",
						"comment":      "initial",
						"disable":      "false",
						"origin":       "user-created",
					}),
				),
			},
			{
				Config: te.RenderConfig(`
					resource "proxmox_notification_endpoint_sendmail" "acc_basic" {
						name         = "acc-sendmail-basic"
						mailto       = ["ops@example.com", "oncall@example.com"]
						mailto_user  = ["root@pam"]
						from_address = "alerts@example.com"
						author       = "Proxmox VE"
						comment      = "updated"
						disable      = true
					}`),
				Check: resource.ComposeTestCheckFunc(
					test.ResourceAttributes("proxmox_notification_endpoint_sendmail.acc_basic", map[string]string{
						"mailto.#":      "2",
						"mailto_user.#": "1",
						"mailto_user.0": "root@pam",
						"comment":       "updated",
						"disable":       "true",
					}),
				),
			},
			{
				Config: te.RenderConfig(`
					resource "proxmox_notification_endpoint_sendmail" "acc_basic" {
						name   = "acc-sendmail-basic"
						mailto = ["ops@example.com"]
					}`),
				Check: resource.ComposeTestCheckFunc(
					test.ResourceAttributes("proxmox_notification_endpoint_sendmail.acc_basic", map[string]string{
						"mailto.#": "1",
						"disable":  "false",
					}),
					test.NoResourceAttributesSet("proxmox_notification_endpoint_sendmail.acc_basic", []string{
						"from_address",
						"author",
						"comment",
						"mailto_user",
					}),
				),
			},
			{
				ResourceName:      "proxmox_notification_endpoint_sendmail.acc_basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		}},
		{"create with mailto_user only & datasource read", []resource.TestStep{
			{
				Config: te.RenderConfig(`
					resource "proxmox_notification_endpoint_sendmail" "acc_user" {
						name        = "acc-sendmail-user"
						mailto_user = ["root@pam"]
						comment     = "user-only"
					}
					data "proxmox_notification_endpoint_sendmail" "acc_user" {
						name = proxmox_notification_endpoint_sendmail.acc_user.name
					}`),
				Check: resource.ComposeTestCheckFunc(
					test.ResourceAttributes("data.proxmox_notification_endpoint_sendmail.acc_user", map[string]string{
						"id":            "acc-sendmail-user",
						"name":          "acc-sendmail-user",
						"mailto_user.#": "1",
						"mailto_user.0": "root@pam",
						"comment":       "user-only",
						"disable":       "false",
						"origin":        "user-created",
					}),
				),
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProtoV6ProviderFactories: te.AccProviders,
				Steps:                    tt.steps,
			})
		})
	}
}
