/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"github.com/bpg/terraform-provider-proxmox/proxmox/types"
)

// SendmailEndpointData contains the data of a sendmail notification endpoint.
type SendmailEndpointData struct {
	Name        string            `json:"name,omitempty"         url:"name,omitempty"`
	Mailto      []string          `json:"mailto,omitempty"       url:"mailto,omitempty"`
	MailtoUser  []string          `json:"mailto-user,omitempty"  url:"mailto-user,omitempty"`
	FromAddress *string           `json:"from-address,omitempty" url:"from-address,omitempty"`
	Author      *string           `json:"author,omitempty"       url:"author,omitempty"`
	Comment     *string           `json:"comment,omitempty"      url:"comment,omitempty"`
	Disable     *types.CustomBool `json:"disable,omitempty"      url:"disable,omitempty,int"`
	// Origin is response-only: indicates "user-created", "builtin", or "modified-builtin".
	Origin *string `json:"origin,omitempty" url:"-"`
	Digest *string `json:"digest,omitempty" url:"digest,omitempty"`
}

// SendmailEndpointResponseBody contains the body from a single sendmail endpoint API response.
type SendmailEndpointResponseBody struct {
	Data *SendmailEndpointData `json:"data,omitempty"`
}

// SendmailEndpointsResponseBody contains the body from a list-sendmail-endpoints API response.
type SendmailEndpointsResponseBody struct {
	Data *[]SendmailEndpointData `json:"data,omitempty"`
}

// SendmailEndpointRequestData contains the data for a sendmail endpoint POST/PUT request.
type SendmailEndpointRequestData struct {
	SendmailEndpointData

	Delete []string `url:"delete,omitempty,comma"`
}
