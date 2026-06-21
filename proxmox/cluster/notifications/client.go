/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package notifications provides a client for the Proxmox VE cluster notifications API
// (/cluster/notifications/*), introduced in PVE 8.1 and expanded in 8.2/8.3.
package notifications

import (
	"fmt"

	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
)

// Client is an interface for accessing the Proxmox cluster notifications API.
type Client struct {
	api.Client
}

// ExpandPath expands a relative path to the cluster notifications API path.
func (c *Client) ExpandPath(path string) string {
	return fmt.Sprintf("cluster/notifications/%s", path)
}
