/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/config"
	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

var (
	_ datasource.DataSource              = &sendmailEndpointDatasource{}
	_ datasource.DataSourceWithConfigure = &sendmailEndpointDatasource{}
)

type sendmailEndpointDatasource struct {
	client *notifications.Client
}

// NewSendmailEndpointDatasource creates a new sendmail notification endpoint data source.
func NewSendmailEndpointDatasource() datasource.DataSource {
	return &sendmailEndpointDatasource{}
}

func (d *sendmailEndpointDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "proxmox_notification_endpoint_sendmail"
}

func (d *sendmailEndpointDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(config.DataSource)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Datasource Configure Type",
			fmt.Sprintf("Expected config.DataSource, got: %T", req.ProviderData),
		)

		return
	}

	d.client = cfg.Client.Cluster().Notifications()
}

func (d *sendmailEndpointDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Proxmox VE sendmail notification endpoint by name. Requires `Mapping.Audit` " +
			"or `Mapping.Modify` on `/mapping/notifications`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of this resource (same as `name`).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the sendmail endpoint to look up.",
				Required:    true,
			},
			"mailto": schema.SetAttribute{
				Description: "List of email addresses to deliver notifications to.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"mailto_user": schema.SetAttribute{
				Description: "List of PVE user IDs whose configured email addresses receive notifications.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"from_address": schema.StringAttribute{
				Description: "Sender (`From:`) address used by the local sendmail.",
				Computed:    true,
			},
			"author": schema.StringAttribute{
				Description: "Author name placed in outgoing mails.",
				Computed:    true,
			},
			"comment": schema.StringAttribute{
				Description: "Free-form comment.",
				Computed:    true,
			},
			"disable": schema.BoolAttribute{
				Description: "Whether the endpoint is disabled.",
				Computed:    true,
			},
			"origin": schema.StringAttribute{
				Description: "Origin of this entry: `user-created`, `builtin`, or `modified-builtin`.",
				Computed:    true,
			},
		},
	}
}

//nolint:dupl // sendmail and matcher datasource Read methods share the same shape; distinct PVE resources.
func (d *sendmailEndpointDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state sendmailEndpointDatasourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	data, err := d.client.GetSendmailEndpoint(ctx, name)
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Sendmail Endpoint %q Not Found", name),
				err.Error(),
			)

			return
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Sendmail Endpoint %q", name),
			err.Error(),
		)

		return
	}

	readModel := &sendmailEndpointDatasourceModel{}
	readModel.fromAPI(ctx, name, data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}
