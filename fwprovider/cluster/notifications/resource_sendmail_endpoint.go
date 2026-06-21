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

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/fwprovider/config"
	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

var (
	_ resource.Resource                = &sendmailEndpointResource{}
	_ resource.ResourceWithConfigure   = &sendmailEndpointResource{}
	_ resource.ResourceWithImportState = &sendmailEndpointResource{}
)

type sendmailEndpointResource struct {
	client *notifications.Client
}

// NewSendmailEndpointResource creates a new sendmail notification endpoint resource.
func NewSendmailEndpointResource() resource.Resource {
	return &sendmailEndpointResource{}
}

func (r *sendmailEndpointResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "proxmox_notification_endpoint_sendmail"
}

func (r *sendmailEndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sendmailEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Proxmox VE sendmail notification endpoint " +
			"(`/cluster/notifications/endpoints/sendmail`). Requires `Mapping.Modify` on " +
			"`/mapping/notifications`; create/update additionally require one of " +
			"`Sys.Audit`, `Sys.Modify`, or `Sys.AccessNetwork` on `/`. Available since PVE 8.1.",
		Attributes: map[string]schema.Attribute{
			"id": attribute.ResourceID(),
			"name": schema.StringAttribute{
				Description: "The unique name of the sendmail endpoint. Used as the resource ID in PVE.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mailto": schema.SetAttribute{
				Description: "List of email addresses to deliver notifications to.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"mailto_user": schema.SetAttribute{
				Description: "List of PVE user IDs whose configured email addresses receive notifications.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"from_address": schema.StringAttribute{
				Description: "Sender (`From:`) address used by the local sendmail.",
				Optional:    true,
			},
			"author": schema.StringAttribute{
				Description: "Author name placed in outgoing mails.",
				Optional:    true,
			},
			"comment": schema.StringAttribute{
				Description: "Free-form comment.",
				Optional:    true,
			},
			"disable": schema.BoolAttribute{
				Description: "Whether the endpoint is disabled. Defaults to `false`.",
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

func (r *sendmailEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sendmailEndpointModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	if err := r.client.CreateSendmailEndpoint(ctx, body); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Create Sendmail Endpoint %q", name),
			err.Error(),
		)

		return
	}

	data, err := r.client.GetSendmailEndpoint(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Sendmail Endpoint %q After Creation", name),
			err.Error(),
		)

		return
	}

	readModel := &sendmailEndpointModel{}
	readModel.fromAPI(ctx, name, data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *sendmailEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sendmailEndpointModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	data, err := r.client.GetSendmailEndpoint(ctx, name)
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Sendmail Endpoint %q", name),
			err.Error(),
		)

		return
	}

	readModel := &sendmailEndpointModel{}
	readModel.fromAPI(ctx, name, data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

//nolint:dupl // sendmail and matcher Update share the same delete-then-PUT-then-refresh shape; distinct PVE resources.
func (r *sendmailEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state sendmailEndpointModel

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

	attribute.CheckDelete(plan.Mailto, state.Mailto, &toDelete, "mailto")
	attribute.CheckDelete(plan.MailtoUser, state.MailtoUser, &toDelete, "mailto-user")
	attribute.CheckDelete(plan.FromAddress, state.FromAddress, &toDelete, "from-address")
	attribute.CheckDelete(plan.Author, state.Author, &toDelete, "author")
	attribute.CheckDelete(plan.Comment, state.Comment, &toDelete, "comment")

	body.Delete = toDelete
	body.Digest = attribute.StringPtrFromValue(state.Digest)

	name := plan.Name.ValueString()

	if err := r.client.UpdateSendmailEndpoint(ctx, body); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Update Sendmail Endpoint %q", name),
			err.Error(),
		)

		return
	}

	data, err := r.client.GetSendmailEndpoint(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Sendmail Endpoint %q After Update", name),
			err.Error(),
		)

		return
	}

	readModel := &sendmailEndpointModel{}
	readModel.fromAPI(ctx, name, data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *sendmailEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sendmailEndpointModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	if err := r.client.DeleteSendmailEndpoint(ctx, name); err != nil &&
		!errors.Is(err, api.ErrResourceDoesNotExist) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Delete Sendmail Endpoint %q", name),
			err.Error(),
		)
	}
}

func (r *sendmailEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	data, err := r.client.GetSendmailEndpoint(ctx, req.ID)
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Sendmail Endpoint %q Not Found", req.ID),
				err.Error(),
			)

			return
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Import Sendmail Endpoint %q", req.ID),
			err.Error(),
		)

		return
	}

	readModel := &sendmailEndpointModel{}
	readModel.fromAPI(ctx, req.ID, data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}
