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
	_ datasource.DataSource              = &matcherDatasource{}
	_ datasource.DataSourceWithConfigure = &matcherDatasource{}
)

type matcherDatasource struct {
	client *notifications.Client
}

// NewMatcherDatasource creates a new notification matcher data source.
func NewMatcherDatasource() datasource.DataSource {
	return &matcherDatasource{}
}

func (d *matcherDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "proxmox_notification_matcher"
}

func (d *matcherDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *matcherDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Proxmox VE notification matcher by name. Requires `Mapping.Audit` " +
			"or `Mapping.Modify` on `/mapping/notifications`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of this resource (same as `name`).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the matcher to look up.",
				Required:    true,
			},
			"match_field": schema.SetAttribute{
				Description: "Notification metadata field matchers, formatted as `(exact|regex):<field>=<value>`.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"match_severity": schema.SetAttribute{
				Description: "Notification severities matched by this matcher.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"match_calendar": schema.SetAttribute{
				Description: "Calendar event specs matched by this matcher.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"target": schema.SetAttribute{
				Description: "Notification targets invoked when this matcher matches.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"mode": schema.StringAttribute{
				Description: "How multiple match conditions are combined: `all` (AND) or `any` (OR).",
				Computed:    true,
			},
			"invert_match": schema.BoolAttribute{
				Description: "Whether the result of the entire matcher is inverted.",
				Computed:    true,
			},
			"comment": schema.StringAttribute{
				Description: "Free-form comment.",
				Computed:    true,
			},
			"disable": schema.BoolAttribute{
				Description: "Whether the matcher is disabled.",
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
func (d *matcherDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state matcherDatasourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	data, err := d.client.GetMatcher(ctx, name)
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Notification Matcher %q Not Found", name),
				err.Error(),
			)

			return
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Notification Matcher %q", name),
			err.Error(),
		)

		return
	}

	readModel := &matcherDatasourceModel{}
	readModel.fromAPI(ctx, name, data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}
