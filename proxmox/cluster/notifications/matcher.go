/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

//nolint:dupl // CRUD shape mirrors sendmail.go; targets a distinct PVE resource
package notifications

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
)

// GetMatcher retrieves a notification matcher by name.
func (c *Client) GetMatcher(ctx context.Context, name string) (*MatcherData, error) {
	resBody := &MatcherResponseBody{}

	err := c.DoRequest(ctx, http.MethodGet, c.ExpandPath("matchers/"+name), nil, resBody)
	if err != nil {
		return nil, fmt.Errorf("error reading notification matcher: %w", err)
	}

	if resBody.Data == nil {
		return nil, api.ErrNoDataObjectInResponse
	}

	return resBody.Data, nil
}

// ListMatchers retrieves all notification matchers.
func (c *Client) ListMatchers(ctx context.Context) (*[]MatcherData, error) {
	resBody := &MatchersResponseBody{}

	err := c.DoRequest(ctx, http.MethodGet, c.ExpandPath("matchers"), nil, resBody)
	if err != nil {
		return nil, fmt.Errorf("error listing notification matchers: %w", err)
	}

	if resBody.Data == nil {
		return nil, api.ErrNoDataObjectInResponse
	}

	return resBody.Data, nil
}

// CreateMatcher creates a new notification matcher.
func (c *Client) CreateMatcher(ctx context.Context, data *MatcherRequestData) error {
	err := c.DoRequest(ctx, http.MethodPost, c.ExpandPath("matchers"), data, nil)
	if err != nil {
		return fmt.Errorf("error creating notification matcher: %w", err)
	}

	return nil
}

// UpdateMatcher updates an existing notification matcher.
func (c *Client) UpdateMatcher(ctx context.Context, data *MatcherRequestData) error {
	err := c.DoRequest(ctx, http.MethodPut, c.ExpandPath("matchers/"+data.Name), data, nil)
	if err != nil {
		return fmt.Errorf("error updating notification matcher: %w", err)
	}

	return nil
}

// DeleteMatcher deletes a notification matcher by name.
func (c *Client) DeleteMatcher(ctx context.Context, name string) error {
	err := c.DoRequest(ctx, http.MethodDelete, c.ExpandPath("matchers/"+name), nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting notification matcher: %w", err)
	}

	return nil
}
