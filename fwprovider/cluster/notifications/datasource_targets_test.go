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

func TestAccDataSourceNotificationTargets(t *testing.T) {
	te := test.InitEnvironment(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: te.AccProviders,
		Steps: []resource.TestStep{
			{
				Config: te.RenderConfig(`
					resource "proxmox_notification_endpoint_sendmail" "acc_listed" {
						name    = "acc-targets-listed"
						mailto  = ["ops@example.com"]
						comment = "listed in targets aggregate"
					}
					data "proxmox_notification_targets" "acc_all" {
						depends_on = [proxmox_notification_endpoint_sendmail.acc_listed]
					}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.proxmox_notification_targets.acc_all", "id", "targets",
					),
					resource.TestCheckResourceAttrSet(
						"data.proxmox_notification_targets.acc_all", "targets.#",
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.proxmox_notification_targets.acc_all", "targets.*",
						map[string]string{
							"name":   "acc-targets-listed",
							"type":   "sendmail",
							"origin": "user-created",
						},
					),
				),
			},
		},
	})
}
