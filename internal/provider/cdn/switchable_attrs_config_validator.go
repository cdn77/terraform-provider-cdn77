package cdn

import (
	"context"
	"fmt"
	"slices"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	_ datasource.ConfigValidator = &SwitchableAttrsConfigValidator{}
	_ resource.ConfigValidator   = &SwitchableAttrsConfigValidator{}
)

type cdnSwitchableAttribute struct {
	attr                           string
	switchAttr                     string
	switchValue                    any
	switchValueIsUnknown           bool
	switchDisabledValues           []any
	controlledAttr                 string
	controlledValueIsNullOrUnknown bool
}

type SwitchableAttrsConfigValidator struct{}

func NewNullableListsConfigValidator() *SwitchableAttrsConfigValidator {
	return &SwitchableAttrsConfigValidator{}
}

func (v SwitchableAttrsConfigValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (SwitchableAttrsConfigValidator) MarkdownDescription(_ context.Context) string {
	return "Checks that nested attributes with switchable attribute have the controlled attributes set to null"
}

func (v SwitchableAttrsConfigValidator) ValidateDataSource(
	ctx context.Context,
	req datasource.ValidateConfigRequest,
	resp *datasource.ValidateConfigResponse,
) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v SwitchableAttrsConfigValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v SwitchableAttrsConfigValidator) Validate(ctx context.Context, config tfsdk.Config) (diags diag.Diagnostics) {
	var data Model

	if diags.Append(config.Get(ctx, &data)...); diags.HasError() {
		return diags
	}

	for _, switchableAttribute := range v.getSwitchableAttributes(data) {
		if switchableAttribute.switchValueIsUnknown {
			continue
		}

		if !slices.Contains(switchableAttribute.switchDisabledValues, switchableAttribute.switchValue) {
			continue
		}

		if switchableAttribute.controlledValueIsNullOrUnknown {
			continue
		}

		attrPath := path.Root(switchableAttribute.attr)
		controlledAttrPath := attrPath.AtName(switchableAttribute.controlledAttr)

		var switchValue string
		switch value := switchableAttribute.switchValue.(type) {
		case string:
			switchValue = value
		case bool:
			switchValue = fmt.Sprintf("%t", value)
		default:
			panic(fmt.Sprintf("unknown type %T", switchableAttribute.switchValue))
		}

		diags.Append(
			validatordiag.InvalidAttributeCombinationDiagnostic(
				controlledAttrPath,
				fmt.Sprintf(
					`Attribute "%s" mustn't be set when attribute "%s" is set to "%s"`,
					controlledAttrPath,
					attrPath.AtName(switchableAttribute.switchAttr),
					switchValue,
				),
			),
		)
	}

	return diags
}

func (SwitchableAttrsConfigValidator) getSwitchableAttributes(data Model) []cdnSwitchableAttribute {
	var switchableAttributes []cdnSwitchableAttribute

	if data.GeoProtection != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                           "geo_protection",
			switchAttr:                     "type",
			switchValue:                    data.GeoProtection.Type.ValueString(),
			switchValueIsUnknown:           data.GeoProtection.Type.IsUnknown(),
			switchDisabledValues:           []any{string(cdn77.Disabled)},
			controlledAttr:                 "countries",
			controlledValueIsNullOrUnknown: util.IsNullOrUnknown(data.GeoProtection.Countries),
		})
	}

	if data.HotlinkProtection != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                           "hotlink_protection",
			switchAttr:                     "type",
			switchValue:                    data.HotlinkProtection.Type.ValueString(),
			switchValueIsUnknown:           data.HotlinkProtection.Type.IsUnknown(),
			switchDisabledValues:           []any{string(cdn77.Disabled)},
			controlledAttr:                 "domains",
			controlledValueIsNullOrUnknown: util.IsNullOrUnknown(data.HotlinkProtection.Domains),
		})
	}

	if data.HttpsRedirect != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                           "https_redirect",
			switchAttr:                     "enabled",
			switchValue:                    data.HttpsRedirect.Enabled.ValueBool(),
			switchValueIsUnknown:           data.HttpsRedirect.Enabled.IsUnknown(),
			switchDisabledValues:           []any{false},
			controlledAttr:                 "code",
			controlledValueIsNullOrUnknown: util.IsNullOrUnknown(data.HttpsRedirect.Code),
		})
	}

	if data.IpProtection != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                           "ip_protection",
			switchAttr:                     "type",
			switchValue:                    data.IpProtection.Type.ValueString(),
			switchValueIsUnknown:           data.IpProtection.Type.IsUnknown(),
			switchDisabledValues:           []any{string(cdn77.Disabled)},
			controlledAttr:                 "ips",
			controlledValueIsNullOrUnknown: util.IsNullOrUnknown(data.IpProtection.Ips),
		})
	}

	if data.QueryString != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                 "query_string",
			switchAttr:           "ignore_type",
			switchValue:          data.QueryString.IgnoreType.ValueString(),
			switchValueIsUnknown: data.QueryString.IgnoreType.IsUnknown(),
			switchDisabledValues: []any{
				string(cdn77.QueryStringIgnoreTypeNone),
				string(cdn77.QueryStringIgnoreTypeAll),
			},
			controlledAttr:                 "parameters",
			controlledValueIsNullOrUnknown: util.IsNullOrUnknown(data.QueryString.Parameters),
		})
	}

	if data.SecureToken != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                           "secure_token",
			switchAttr:                     "type",
			switchValue:                    data.SecureToken.Type.ValueString(),
			switchValueIsUnknown:           data.SecureToken.Type.IsUnknown(),
			switchDisabledValues:           []any{string(cdn77.SecureTokenTypeNone)},
			controlledAttr:                 "token",
			controlledValueIsNullOrUnknown: util.IsNullOrUnknown(data.SecureToken.Token),
		})
	}

	if data.Ssl != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                           "ssl",
			switchAttr:                     "type",
			switchValue:                    data.Ssl.Type.ValueString(),
			switchValueIsUnknown:           data.Ssl.Type.IsUnknown(),
			switchDisabledValues:           []any{string(cdn77.InstantSsl), string(cdn77.None)},
			controlledAttr:                 "ssl_id",
			controlledValueIsNullOrUnknown: util.IsNullOrUnknown(data.Ssl.SslId),
		})
	}

	return switchableAttributes
}
