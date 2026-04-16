package cdn_test

import (
	"testing"

	"github.com/cdn77/terraform-provider-cdn77/internal/provider/cdn"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSwitchableAttrsConfigValidator_HttpsRedirectUnknownValues(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	s := cdn.CreateResourceSchema()

	rootObjType, ok := s.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("expected schema type to be tftypes.Object, got %T", s.Type().TerraformType(ctx))
	}

	httpsRedirectObjType, ok := rootObjType.AttributeTypes["https_redirect"].(tftypes.Object)
	if !ok {
		t.Fatalf("expected https_redirect to be tftypes.Object")
	}

	buildConfig := func(enabledVal, codeVal tftypes.Value) tfsdk.Config {
		attrVals := make(map[string]tftypes.Value, len(rootObjType.AttributeTypes))
		for name, typ := range rootObjType.AttributeTypes {
			attrVals[name] = tftypes.NewValue(typ, nil)
		}

		attrVals["https_redirect"] = tftypes.NewValue(httpsRedirectObjType, map[string]tftypes.Value{
			"enabled": enabledVal,
			"code":    codeVal,
		})

		return tfsdk.Config{
			Raw:    tftypes.NewValue(rootObjType, attrVals),
			Schema: s,
		}
	}

	tests := []struct {
		name        string
		enabled     tftypes.Value
		code        tftypes.Value
		expectError bool
	}{
		{
			name:        "enabled=false, code=null",
			enabled:     tftypes.NewValue(tftypes.Bool, false),
			code:        tftypes.NewValue(tftypes.Number, nil),
			expectError: false,
		},
		{
			name:        "enabled=false, code=unknown",
			enabled:     tftypes.NewValue(tftypes.Bool, false),
			code:        tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			expectError: false,
		},
		{
			name:        "enabled=unknown, code=301",
			enabled:     tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			code:        tftypes.NewValue(tftypes.Number, 301),
			expectError: false,
		},
		{
			name:        "enabled=false, code=301",
			enabled:     tftypes.NewValue(tftypes.Bool, false),
			code:        tftypes.NewValue(tftypes.Number, 301),
			expectError: true,
		},
		{
			name:        "enabled=true, code=301",
			enabled:     tftypes.NewValue(tftypes.Bool, true),
			code:        tftypes.NewValue(tftypes.Number, 301),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := cdn.NewNullableListsConfigValidator()
			diags := v.Validate(ctx, buildConfig(tt.enabled, tt.code))

			if tt.expectError && !diags.HasError() {
				t.Error("expected a validation error but got none")
			}

			if !tt.expectError && diags.HasError() {
				t.Errorf("expected no validation error but got: %s", diags)
			}
		})
	}
}
