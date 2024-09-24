package util

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func StringSetToSlice(ctx context.Context, diags *diag.Diagnostics, attrPath path.Path, s types.Set) ([]string, bool) {
	values := make([]types.String, 0, len(s.Elements()))
	if diags.Append(s.ElementsAs(ctx, &values, false)...); diags.HasError() {
		return nil, false
	}

	valuesRaw := make([]string, len(values))

	for i, v := range values {
		if v.IsNull() {
			diags.AddAttributeError(
				attrPath,
				"Set contains null value",
				fmt.Sprintf(`Attribute "%s" contains a null value, which is not allowed`, attrPath),
			)

			return nil, false
		}

		if v.IsUnknown() {
			diags.AddAttributeError(
				attrPath,
				"Set contains unknown value",
				fmt.Sprintf(`Attribute "%s" contains an unknown value, which is not allowed`, attrPath),
			)

			return nil, false
		}

		valuesRaw[i] = v.ValueString()
	}

	return valuesRaw, true
}

func StringMapToMap(
	ctx context.Context,
	diags *diag.Diagnostics,
	attrPath path.Path,
	m types.Map,
) (map[string]string, bool) {
	tfStringMap := make(map[string]types.String, len(m.Elements()))
	if diags.Append(m.ElementsAs(ctx, &tfStringMap, false)...); diags.HasError() {
		return nil, false
	}

	stringMap := make(map[string]string, len(tfStringMap))

	for key, v := range tfStringMap {
		if v.IsNull() {
			diags.AddAttributeError(
				attrPath.AtMapKey(key),
				"Map contains null value",
				fmt.Sprintf(
					`Attribute "%s" contains a null value under key "%s", which is not allowed`, attrPath, key,
				),
			)

			return nil, false
		}

		if v.IsUnknown() {
			diags.AddAttributeError(
				attrPath.AtMapKey(key),
				"Map contains unknown value",
				fmt.Sprintf(
					`Attribute "%s" contains an unknown value under key "%s", which is not allowed`, attrPath, key,
				),
			)

			return nil, false
		}

		stringMap[key] = v.ValueString()
	}

	return stringMap, true
}

func SetValueFrom(ctx context.Context, diags *diag.Diagnostics, elemType attr.Type, elements any) basetypes.SetValue {
	set, ds := types.SetValueFrom(ctx, elemType, elements)
	if ds != nil {
		diags.Append(ds...)
	}

	return set
}

func MapValueFrom(ctx context.Context, diags *diag.Diagnostics, elemType attr.Type, elements any) basetypes.MapValue {
	m, ds := types.MapValueFrom(ctx, elemType, elements)
	if ds != nil {
		diags.Append(ds...)
	}

	return m
}
