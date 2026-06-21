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

type matcherModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	MatchField    types.Set    `tfsdk:"match_field"`
	MatchSeverity types.Set    `tfsdk:"match_severity"`
	MatchCalendar types.Set    `tfsdk:"match_calendar"`
	Target        types.Set    `tfsdk:"target"`
	Mode          types.String `tfsdk:"mode"`
	InvertMatch   types.Bool   `tfsdk:"invert_match"`
	Comment       types.String `tfsdk:"comment"`
	Disable       types.Bool   `tfsdk:"disable"`
	Origin        types.String `tfsdk:"origin"`
	Digest        types.String `tfsdk:"digest"`
}

func (m *matcherModel) fromAPI(ctx context.Context, name string, data *notifications.MatcherData, diags *diag.Diagnostics) {
	m.ID = types.StringValue(name)
	m.Name = types.StringValue(name)
	m.MatchField = stringSliceToSet(ctx, data.MatchField, diags)
	m.MatchSeverity = stringSliceToSet(ctx, data.MatchSeverity, diags)
	m.MatchCalendar = stringSliceToSet(ctx, data.MatchCalendar, diags)
	m.Target = stringSliceToSet(ctx, data.Target, diags)
	// `mode` defaults to "all" server-side when unset; reflect that in state.
	if data.Mode == nil {
		m.Mode = types.StringValue("all")
	} else {
		m.Mode = types.StringPointerValue(data.Mode)
	}

	m.InvertMatch = boolOrFalse(data.InvertMatch)
	m.Comment = types.StringPointerValue(data.Comment)
	m.Disable = boolOrFalse(data.Disable)
	m.Origin = types.StringPointerValue(data.Origin)
	m.Digest = types.StringPointerValue(data.Digest)
}

func (m *matcherModel) toAPI(ctx context.Context, diags *diag.Diagnostics) *notifications.MatcherRequestData {
	data := &notifications.MatcherRequestData{}
	data.Name = m.Name.ValueString()
	data.MatchField = setToStringSlice(ctx, m.MatchField, diags)
	data.MatchSeverity = setToStringSlice(ctx, m.MatchSeverity, diags)
	data.MatchCalendar = setToStringSlice(ctx, m.MatchCalendar, diags)
	data.Target = setToStringSlice(ctx, m.Target, diags)
	data.Mode = attribute.StringPtrFromValue(m.Mode)
	data.InvertMatch = attribute.CustomBoolPtrFromValue(m.InvertMatch)
	data.Comment = attribute.StringPtrFromValue(m.Comment)
	data.Disable = attribute.CustomBoolPtrFromValue(m.Disable)

	return data
}

type matcherDatasourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	MatchField    types.Set    `tfsdk:"match_field"`
	MatchSeverity types.Set    `tfsdk:"match_severity"`
	MatchCalendar types.Set    `tfsdk:"match_calendar"`
	Target        types.Set    `tfsdk:"target"`
	Mode          types.String `tfsdk:"mode"`
	InvertMatch   types.Bool   `tfsdk:"invert_match"`
	Comment       types.String `tfsdk:"comment"`
	Disable       types.Bool   `tfsdk:"disable"`
	Origin        types.String `tfsdk:"origin"`
}

func (m *matcherDatasourceModel) fromAPI(ctx context.Context, name string, data *notifications.MatcherData, diags *diag.Diagnostics) {
	m.ID = types.StringValue(name)
	m.Name = types.StringValue(name)
	m.MatchField = stringSliceToSet(ctx, data.MatchField, diags)
	m.MatchSeverity = stringSliceToSet(ctx, data.MatchSeverity, diags)
	m.MatchCalendar = stringSliceToSet(ctx, data.MatchCalendar, diags)
	m.Target = stringSliceToSet(ctx, data.Target, diags)

	if data.Mode == nil {
		m.Mode = types.StringValue("all")
	} else {
		m.Mode = types.StringPointerValue(data.Mode)
	}

	m.InvertMatch = boolOrFalse(data.InvertMatch)
	m.Comment = attribute.StringValueFromPtr(data.Comment)
	m.Disable = boolOrFalse(data.Disable)
	m.Origin = attribute.StringValueFromPtr(data.Origin)
}
