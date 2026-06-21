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

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/fwprovider/config"
	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

var (
	_ resource.Resource                = &matcherResource{}
	_ resource.ResourceWithConfigure   = &matcherResource{}
	_ resource.ResourceWithImportState = &matcherResource{}
)

type matcherResource struct {
	client *notifications.Client
}

// NewMatcherResource creates a new notification matcher resource.
func NewMatcherResource() resource.Resource {
	return &matcherResource{}
}

func (r *matcherResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "proxmox_notification_matcher"
}

func (r *matcherResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(config.Resource)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected config.Resource, got: %T", req.ProviderData),
		)

		return
	}

	r.client = cfg.Client.Cluster().Notifications()
}

func (r *matcherResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Proxmox VE notification matcher (`/cluster/notifications/matchers`). " +
			"Matchers route notifications to one or more targets based on metadata, severity, and calendar " +
			"conditions. Requires `Mapping.Modify` on `/mapping/notifications`. Available since PVE 8.1.",
		Attributes: map[string]schema.Attribute{
			"id": attribute.ResourceID(),
			"name": schema.StringAttribute{
				Description: "The unique name of the matcher. Used as the resource ID in PVE.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"match_field": schema.SetAttribute{
				Description: "Match notification metadata fields. Each entry is `(exact|regex):<field>=<value>` " +
					"(e.g., `exact:type=vzdump`).",
				ElementType: types.StringType,
				Optional:    true,
			},
			"match_severity": schema.SetAttribute{
				Description: "Match notification severities (e.g., `info`, `notice`, `warning`, `error`, `unknown`).",
				ElementType: types.StringType,
				Optional:    true,
			},
			"match_calendar": schema.SetAttribute{
				Description: "Match notification timestamps. Each entry is a systemd-style calendar event spec.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"target": schema.SetAttribute{
				Description: "Names of notification targets to invoke when this matcher matches.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"mode": schema.StringAttribute{
				Description: "How multiple match conditions are combined: `all` (AND) or `any` (OR). " +
					"Defaults to `all`.",
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString("all"),
				Validators: []validator.String{stringvalidator.OneOf("all", "any")},
			},
			"invert_match": schema.BoolAttribute{
				Description: "Whether to invert the result of the entire matcher. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"comment": schema.StringAttribute{
				Description: "Free-form comment.",
				Optional:    true,
			},
			"disable": schema.BoolAttribute{
				Description: "Whether the matcher is disabled. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"origin": schema.StringAttribute{
				Description: "Origin of this entry: `user-created`, `builtin`, or `modified-builtin`.",
				Computed:    true,
			},
			"digest": schema.StringAttribute{
				Description: "SHA-1 digest of the notifications configuration, used for optimistic locking on update.",
				Computed:    true,
			},
		},
	}
}

func (r *matcherResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan matcherModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	if err := r.client.CreateMatcher(ctx, body); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Create Notification Matcher %q", name),
			err.Error(),
		)

		return
	}

	data, err := r.client.GetMatcher(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Notification Matcher %q After Creation", name),
			err.Error(),
		)

		return
	}

	readModel := &matcherModel{}
	readModel.fromAPI(ctx, name, data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *matcherResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state matcherModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	data, err := r.client.GetMatcher(ctx, name)
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Notification Matcher %q", name),
			err.Error(),
		)

		return
	}

	readModel := &matcherModel{}
	readModel.fromAPI(ctx, name, data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

//nolint:dupl // sendmail and matcher Update share the same delete-then-PUT-then-refresh shape; distinct PVE resources.
func (r *matcherResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state matcherModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var toDelete []string

	attribute.CheckDelete(plan.MatchField, state.MatchField, &toDelete, "match-field")
	attribute.CheckDelete(plan.MatchSeverity, state.MatchSeverity, &toDelete, "match-severity")
	attribute.CheckDelete(plan.MatchCalendar, state.MatchCalendar, &toDelete, "match-calendar")
	attribute.CheckDelete(plan.Target, state.Target, &toDelete, "target")
	attribute.CheckDelete(plan.Comment, state.Comment, &toDelete, "comment")

	body.Delete = toDelete
	body.Digest = attribute.StringPtrFromValue(state.Digest)

	name := plan.Name.ValueString()

	if err := r.client.UpdateMatcher(ctx, body); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Update Notification Matcher %q", name),
			err.Error(),
		)

		return
	}

	data, err := r.client.GetMatcher(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Notification Matcher %q After Update", name),
			err.Error(),
		)

		return
	}

	readModel := &matcherModel{}
	readModel.fromAPI(ctx, name, data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *matcherResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state matcherModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	if err := r.client.DeleteMatcher(ctx, name); err != nil &&
		!errors.Is(err, api.ErrResourceDoesNotExist) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Delete Notification Matcher %q", name),
			err.Error(),
		)
	}
}

func (r *matcherResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	data, err := r.client.GetMatcher(ctx, req.ID)
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Notification Matcher %q Not Found", req.ID),
				err.Error(),
			)

			return
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Import Notification Matcher %q", req.ID),
			err.Error(),
		)

		return
	}

	readModel := &matcherModel{}
	readModel.fromAPI(ctx, req.ID, data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}
