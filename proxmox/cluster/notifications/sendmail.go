/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

//nolint:dupl // CRUD shape mirrors matcher.go; targets a distinct PVE resource
package notifications

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
)

// GetSendmailEndpoint retrieves a sendmail notification endpoint by name.
func (c *Client) GetSendmailEndpoint(ctx context.Context, name string) (*SendmailEndpointData, error) {
	resBody := &SendmailEndpointResponseBody{}

	err := c.DoRequest(ctx, http.MethodGet, c.ExpandPath("endpoints/sendmail/"+name), nil, resBody)
	if err != nil {
		return nil, fmt.Errorf("error reading sendmail endpoint: %w", err)
	}

	if resBody.Data == nil {
		return nil, api.ErrNoDataObjectInResponse
	}

	return resBody.Data, nil
}

// ListSendmailEndpoints retrieves all sendmail notification endpoints.
func (c *Client) ListSendmailEndpoints(ctx context.Context) (*[]SendmailEndpointData, error) {
	resBody := &SendmailEndpointsResponseBody{}

	err := c.DoRequest(ctx, http.MethodGet, c.ExpandPath("endpoints/sendmail"), nil, resBody)
	if err != nil {
		return nil, fmt.Errorf("error listing sendmail endpoints: %w", err)
	}

	if resBody.Data == nil {
		return nil, api.ErrNoDataObjectInResponse
	}

	return resBody.Data, nil
}

// CreateSendmailEndpoint creates a new sendmail notification endpoint.
func (c *Client) CreateSendmailEndpoint(ctx context.Context, data *SendmailEndpointRequestData) error {
	err := c.DoRequest(ctx, http.MethodPost, c.ExpandPath("endpoints/sendmail"), data, nil)
	if err != nil {
		return fmt.Errorf("error creating sendmail endpoint: %w", err)
	}

	return nil
}

// UpdateSendmailEndpoint updates an existing sendmail notification endpoint.
func (c *Client) UpdateSendmailEndpoint(ctx context.Context, data *SendmailEndpointRequestData) error {
	err := c.DoRequest(ctx, http.MethodPut, c.ExpandPath("endpoints/sendmail/"+data.Name), data, nil)
	if err != nil {
		return fmt.Errorf("error updating sendmail endpoint: %w", err)
	}

	return nil
}

// DeleteSendmailEndpoint deletes a sendmail notification endpoint by name.
func (c *Client) DeleteSendmailEndpoint(ctx context.Context, name string) error {
	err := c.DoRequest(ctx, http.MethodDelete, c.ExpandPath("endpoints/sendmail/"+name), nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting sendmail endpoint: %w", err)
	}

	return nil
}
