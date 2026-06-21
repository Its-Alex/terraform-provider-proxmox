/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

type sendmailEndpointModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Mailto      types.Set    `tfsdk:"mailto"`
	MailtoUser  types.Set    `tfsdk:"mailto_user"`
	FromAddress types.String `tfsdk:"from_address"`
	Author      types.String `tfsdk:"author"`
	Comment     types.String `tfsdk:"comment"`
	Disable     types.Bool   `tfsdk:"disable"`
	Origin      types.String `tfsdk:"origin"`
	Digest      types.String `tfsdk:"digest"`
}

func (m *sendmailEndpointModel) fromAPI(ctx context.Context, name string, data *notifications.SendmailEndpointData, diags *diag.Diagnostics) {
	m.ID = types.StringValue(name)
	m.Name = types.StringValue(name)
	m.Mailto = stringSliceToSet(ctx, data.Mailto, diags)
	m.MailtoUser = stringSliceToSet(ctx, data.MailtoUser, diags)
	m.FromAddress = types.StringPointerValue(data.FromAddress)
	m.Author = types.StringPointerValue(data.Author)
	m.Comment = types.StringPointerValue(data.Comment)
	// PVE omits `disable` from GET responses when false; normalize to the schema default.
	m.Disable = boolOrFalse(data.Disable)
	m.Origin = types.StringPointerValue(data.Origin)
	m.Digest = types.StringPointerValue(data.Digest)
}

func (m *sendmailEndpointModel) toAPI(ctx context.Context, diags *diag.Diagnostics) *notifications.SendmailEndpointRequestData {
	data := &notifications.SendmailEndpointRequestData{}
	data.Name = m.Name.ValueString()
	data.Mailto = setToStringSlice(ctx, m.Mailto, diags)
	data.MailtoUser = setToStringSlice(ctx, m.MailtoUser, diags)
	data.FromAddress = attribute.StringPtrFromValue(m.FromAddress)
	data.Author = attribute.StringPtrFromValue(m.Author)
	data.Comment = attribute.StringPtrFromValue(m.Comment)
	data.Disable = attribute.CustomBoolPtrFromValue(m.Disable)

	return data
}

type sendmailEndpointDatasourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Mailto      types.Set    `tfsdk:"mailto"`
	MailtoUser  types.Set    `tfsdk:"mailto_user"`
	FromAddress types.String `tfsdk:"from_address"`
	Author      types.String `tfsdk:"author"`
	Comment     types.String `tfsdk:"comment"`
	Disable     types.Bool   `tfsdk:"disable"`
	Origin      types.String `tfsdk:"origin"`
}

func (m *sendmailEndpointDatasourceModel) fromAPI(
	ctx context.Context,
	name string,
	data *notifications.SendmailEndpointData,
	diags *diag.Diagnostics,
) {
	m.ID = types.StringValue(name)
	m.Name = types.StringValue(name)
	m.Mailto = stringSliceToSet(ctx, data.Mailto, diags)
	m.MailtoUser = stringSliceToSet(ctx, data.MailtoUser, diags)
	m.FromAddress = attribute.StringValueFromPtr(data.FromAddress)
	m.Author = attribute.StringValueFromPtr(data.Author)
	m.Comment = attribute.StringValueFromPtr(data.Comment)
	m.Disable = boolOrFalse(data.Disable)
	m.Origin = attribute.StringValueFromPtr(data.Origin)
}
