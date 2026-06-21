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

	proxmoxtypes "github.com/bpg/terraform-provider-proxmox/proxmox/types"
)

// stringSliceToSet converts a Go []string from a PVE API response to a Terraform
// types.Set of types.String. Returns SetNull when the input is nil so optional
// list fields surface as absent rather than empty in state.
func stringSliceToSet(ctx context.Context, s []string, diags *diag.Diagnostics) types.Set {
	if s == nil {
		return types.SetNull(types.StringType)
	}

	v, d := types.SetValueFrom(ctx, types.StringType, s)
	diags.Append(d...)

	return v
}

// setToStringSlice converts a Terraform types.Set of types.String to a Go []string
// for sending to the PVE API. Returns nil for null/unknown sets so the field is
// omitted from the request via `url:"...,omitempty"`.
func setToStringSlice(ctx context.Context, s types.Set, diags *diag.Diagnostics) []string {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}

	out := make([]string, 0, len(s.Elements()))
	diags.Append(s.ElementsAs(ctx, &out, false)...)

	return out
}

// boolOrFalse unwraps a CustomBool pointer, returning a Terraform false when nil.
// Used for API fields that PVE omits when their value equals the default (false).
func boolOrFalse(b *proxmoxtypes.CustomBool) types.Bool {
	if v := b.PointerBool(); v != nil {
		return types.BoolValue(*v)
	}

	return types.BoolValue(false)
}
