/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/fwprovider/config"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

var (
	_ datasource.DataSource              = &targetsDatasource{}
	_ datasource.DataSourceWithConfigure = &targetsDatasource{}
)

type targetsDatasource struct {
	client *notifications.Client
}

// NewTargetsDatasource creates a new aggregate notification targets data source.
func NewTargetsDatasource() datasource.DataSource {
	return &targetsDatasource{}
}

type targetsDatasourceModel struct {
	ID      types.String `tfsdk:"id"`
	Targets types.List   `tfsdk:"targets"`
}

type targetEntryModel struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Comment types.String `tfsdk:"comment"`
	Disable types.Bool   `tfsdk:"disable"`
	Origin  types.String `tfsdk:"origin"`
}

func targetEntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":    types.StringType,
		"type":    types.StringType,
		"comment": types.StringType,
		"disable": types.BoolType,
		"origin":  types.StringType,
	}
}

func (d *targetsDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "proxmox_notification_targets"
}

func (d *targetsDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *targetsDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns the aggregate list of all configured notification targets across every endpoint " +
			"type (sendmail, smtp, gotify, webhook). Read-only. Requires `Mapping.Audit`, `Mapping.Modify`, " +
			"or `Mapping.Use` on `/mapping/notifications`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of this data source. Always set to `targets`.",
				Computed:    true,
			},
			"targets": schema.ListNestedAttribute{
				Description: "List of every configured notification target.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the target.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Type of the target: `sendmail`, `smtp`, `gotify`, or `webhook`.",
							Computed:    true,
						},
						"comment": schema.StringAttribute{
							Description: "Free-form comment.",
							Computed:    true,
						},
						"disable": schema.BoolAttribute{
							Description: "Whether the target is disabled.",
							Computed:    true,
						},
						"origin": schema.StringAttribute{
							Description: "Origin of this target: `user-created`, `builtin`, or `modified-builtin`.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *targetsDatasource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	data, err := d.client.ListTargets(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to List Notification Targets", err.Error())
		return
	}

	state := targetsDatasourceModel{
		ID:      types.StringValue("targets"),
		Targets: targetsToList(ctx, data, &resp.Diagnostics),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func targetsToList(ctx context.Context, data *[]notifications.TargetData, diags *diag.Diagnostics) types.List {
	objType := types.ObjectType{AttrTypes: targetEntryAttrTypes()}

	if data == nil || len(*data) == 0 {
		return types.ListValueMust(objType, []attr.Value{})
	}

	entries := make([]targetEntryModel, 0, len(*data))

	for _, t := range *data {
		entries = append(entries, targetEntryModel{
			Name:    types.StringValue(t.Name),
			Type:    types.StringValue(t.Type),
			Comment: attribute.StringValueFromPtr(t.Comment),
			Disable: boolOrFalse(t.Disable),
			Origin:  types.StringValue(t.Origin),
		})
	}

	listValue, d := types.ListValueFrom(ctx, objType, entries)
	diags.Append(d...)

	return listValue
}
