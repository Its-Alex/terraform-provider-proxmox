/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
	"github.com/bpg/terraform-provider-proxmox/proxmox/types"
)

// TargetData represents a notification target listing entry. Read-only aggregate
// across all endpoint types (sendmail, gotify, smtp, webhook).
type TargetData struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Comment *string           `json:"comment,omitempty"`
	Disable *types.CustomBool `json:"disable,omitempty"`
	// Origin is one of "user-created", "builtin", or "modified-builtin".
	Origin string `json:"origin"`
}

// TargetsResponseBody contains the body from a list-targets API response.
type TargetsResponseBody struct {
	Data *[]TargetData `json:"data,omitempty"`
}

// ListTargets retrieves the aggregate list of all configured notification targets
// across every endpoint type.
func (c *Client) ListTargets(ctx context.Context) (*[]TargetData, error) {
	resBody := &TargetsResponseBody{}

	err := c.DoRequest(ctx, http.MethodGet, c.ExpandPath("targets"), nil, resBody)
	if err != nil {
		return nil, fmt.Errorf("error listing notification targets: %w", err)
	}

	if resBody.Data == nil {
		return nil, api.ErrNoDataObjectInResponse
	}

	return resBody.Data, nil
}
