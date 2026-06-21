/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"github.com/bpg/terraform-provider-proxmox/proxmox/types"
)

// MatcherData contains the data of a notification matcher.
type MatcherData struct {
	Name          string            `json:"name,omitempty"           url:"name,omitempty"`
	MatchField    []string          `json:"match-field,omitempty"    url:"match-field,omitempty"`
	MatchSeverity []string          `json:"match-severity,omitempty" url:"match-severity,omitempty"`
	MatchCalendar []string          `json:"match-calendar,omitempty" url:"match-calendar,omitempty"`
	Target        []string          `json:"target,omitempty"         url:"target,omitempty"`
	Mode          *string           `json:"mode,omitempty"           url:"mode,omitempty"`
	InvertMatch   *types.CustomBool `json:"invert-match,omitempty"   url:"invert-match,omitempty,int"`
	Comment       *string           `json:"comment,omitempty"        url:"comment,omitempty"`
	Disable       *types.CustomBool `json:"disable,omitempty"        url:"disable,omitempty,int"`
	// Origin is response-only: "user-created", "builtin", or "modified-builtin".
	Origin *string `json:"origin,omitempty" url:"-"`
	Digest *string `json:"digest,omitempty" url:"digest,omitempty"`
}

// MatcherResponseBody contains the body from a single matcher API response.
type MatcherResponseBody struct {
	Data *MatcherData `json:"data,omitempty"`
}

// MatchersResponseBody contains the body from a list-matchers API response.
type MatchersResponseBody struct {
	Data *[]MatcherData `json:"data,omitempty"`
}

// MatcherRequestData contains the data for a matcher POST/PUT request.
type MatcherRequestData struct {
	MatcherData

	Delete []string `url:"delete,omitempty,comma"`
}
