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

func TestAccResourceNotificationMatcher(t *testing.T) {
	te := test.InitEnvironment(t)

	tests := []struct {
		name  string
		steps []resource.TestStep
	}{
		{"create matcher, update, drop optional fields, import", []resource.TestStep{
			{
				Config: te.RenderConfig(`
					resource "proxmox_notification_endpoint_sendmail" "acc_target" {
						name   = "acc-matcher-target"
						mailto = ["ops@example.com"]
					}
					resource "proxmox_notification_matcher" "acc_basic" {
						name           = "acc-matcher-basic"
						target         = [proxmox_notification_endpoint_sendmail.acc_target.name]
						match_severity = ["warning", "error"]
						mode           = "any"
						comment        = "initial"
					}`),
				Check: resource.ComposeTestCheckFunc(
					test.ResourceAttributes("proxmox_notification_matcher.acc_basic", map[string]string{
						"id":               "acc-matcher-basic",
						"name":             "acc-matcher-basic",
						"target.#":         "1",
						"match_severity.#": "2",
						"mode":             "any",
						"invert_match":     "false",
						"disable":          "false",
						"comment":          "initial",
						"origin":           "user-created",
					}),
				),
			},
			{
				Config: te.RenderConfig(`
					resource "proxmox_notification_endpoint_sendmail" "acc_target" {
						name   = "acc-matcher-target"
						mailto = ["ops@example.com"]
					}
					resource "proxmox_notification_matcher" "acc_basic" {
						name           = "acc-matcher-basic"
						target         = [proxmox_notification_endpoint_sendmail.acc_target.name]
						match_severity = ["error"]
						match_field    = ["exact:type=vzdump"]
						match_calendar = ["mon..fri 09:00-17:00"]
						mode           = "all"
						invert_match   = true
						disable        = true
					}`),
				Check: resource.ComposeTestCheckFunc(
					test.ResourceAttributes("proxmox_notification_matcher.acc_basic", map[string]string{
						"match_severity.#": "1",
						"match_severity.0": "error",
						"match_field.#":    "1",
						"match_field.0":    "exact:type=vzdump",
						"match_calendar.#": "1",
						"match_calendar.0": "mon..fri 09:00-17:00",
						"mode":             "all",
						"invert_match":     "true",
						"disable":          "true",
					}),
					test.NoResourceAttributesSet("proxmox_notification_matcher.acc_basic", []string{
						"comment",
					}),
				),
			},
			{
				ResourceName:      "proxmox_notification_matcher.acc_basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		}},
		{"matcher data source by name", []resource.TestStep{
			{
				Config: te.RenderConfig(`
					resource "proxmox_notification_endpoint_sendmail" "acc_ds_target" {
						name   = "acc-matcher-ds-target"
						mailto = ["ops@example.com"]
					}
					resource "proxmox_notification_matcher" "acc_ds" {
						name           = "acc-matcher-ds"
						target         = [proxmox_notification_endpoint_sendmail.acc_ds_target.name]
						match_severity = ["warning"]
					}
					data "proxmox_notification_matcher" "acc_ds" {
						name = proxmox_notification_matcher.acc_ds.name
					}`),
				Check: resource.ComposeTestCheckFunc(
					test.ResourceAttributes("data.proxmox_notification_matcher.acc_ds", map[string]string{
						"id":               "acc-matcher-ds",
						"name":             "acc-matcher-ds",
						"target.#":         "1",
						"match_severity.#": "1",
						"match_severity.0": "warning",
						"mode":             "all",
						"origin":           "user-created",
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
