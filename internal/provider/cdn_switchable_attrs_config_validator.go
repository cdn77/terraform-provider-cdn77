package provider

import (
	"context"
	"fmt"
	"slices"

	"github.com/cdn77/cdn77-client-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	_ datasource.ConfigValidator = &CdnSwitchableAttrsConfigValidator{}
	_ resource.ConfigValidator   = &CdnSwitchableAttrsConfigValidator{}
)

type cdnSwitchableAttribute struct {
	attr                  string
	switchAttr            string
	switchValue           any
	switchDisabledValues  []any
	controlledAttr        string
	controlledValueIsNull bool
}

type CdnSwitchableAttrsConfigValidator struct{}

func NewCdnNullableListsConfigValidator() *CdnSwitchableAttrsConfigValidator {
	return &CdnSwitchableAttrsConfigValidator{}
}

func (v CdnSwitchableAttrsConfigValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (CdnSwitchableAttrsConfigValidator) MarkdownDescription(_ context.Context) string {
	return "Checks that nested attributes with switchable attribute have the controlled attributes set to null"
}

func (v CdnSwitchableAttrsConfigValidator) ValidateDataSource(
	ctx context.Context,
	req datasource.ValidateConfigRequest,
	resp *datasource.ValidateConfigResponse,
) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v CdnSwitchableAttrsConfigValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v CdnSwitchableAttrsConfigValidator) Validate(ctx context.Context, config tfsdk.Config) diag.Diagnostics {
	var data CdnModel
	diags := config.Get(ctx, &data)

	for _, switchableAttribute := range v.getSwitchableAttributes(data) {
		if !slices.Contains(switchableAttribute.switchDisabledValues, switchableAttribute.switchValue) {
			continue
		}

		if switchableAttribute.controlledValueIsNull {
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

func (CdnSwitchableAttrsConfigValidator) getSwitchableAttributes(data CdnModel) []cdnSwitchableAttribute {
	var switchableAttributes []cdnSwitchableAttribute

	if data.SecureToken != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                  "secure_token",
			switchAttr:            "type",
			switchValue:           data.SecureToken.Type.ValueString(),
			switchDisabledValues:  []any{string(cdn77.SecureTokenTypeNone)},
			controlledAttr:        "token",
			controlledValueIsNull: data.SecureToken.Token.IsNull(),
		})
	}

	if data.QueryString != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:        "query_string",
			switchAttr:  "ignore_type",
			switchValue: data.QueryString.IgnoreType.ValueString(),
			switchDisabledValues: []any{
				string(cdn77.QueryStringIgnoreTypeNone),
				string(cdn77.QueryStringIgnoreTypeAll),
			},
			controlledAttr:        "parameters",
			controlledValueIsNull: data.QueryString.Parameters.IsNull(),
		})
	}

	if data.HttpsRedirect != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                  "https_redirect",
			switchAttr:            "enabled",
			switchValue:           data.HttpsRedirect.Enabled.ValueBool(),
			switchDisabledValues:  []any{false},
			controlledAttr:        "code",
			controlledValueIsNull: data.HttpsRedirect.Code.IsNull(),
		})
	}

	if data.Ssl != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                  "ssl",
			switchAttr:            "type",
			switchValue:           data.Ssl.Type.ValueString(),
			switchDisabledValues:  []any{string(cdn77.InstantSsl), string(cdn77.None)},
			controlledAttr:        "ssl_id",
			controlledValueIsNull: data.Ssl.SslId.IsNull(),
		})
	}

	if data.HotlinkProtection != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                  "hotlink_protection",
			switchAttr:            "type",
			switchValue:           data.HotlinkProtection.Type.ValueString(),
			switchDisabledValues:  []any{string(cdn77.Disabled)},
			controlledAttr:        "domains",
			controlledValueIsNull: data.HotlinkProtection.Domains.IsNull(),
		})
	}

	if data.IpProtection != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                  "ip_protection",
			switchAttr:            "type",
			switchValue:           data.IpProtection.Type.ValueString(),
			switchDisabledValues:  []any{string(cdn77.Disabled)},
			controlledAttr:        "ips",
			controlledValueIsNull: data.IpProtection.Ips.IsNull(),
		})
	}

	if data.GeoProtection != nil {
		switchableAttributes = append(switchableAttributes, cdnSwitchableAttribute{
			attr:                  "geo_protection",
			switchAttr:            "type",
			switchValue:           data.GeoProtection.Type.ValueString(),
			switchDisabledValues:  []any{string(cdn77.Disabled)},
			controlledAttr:        "countries",
			controlledValueIsNull: data.GeoProtection.Countries.IsNull(),
		})
	}

	return switchableAttributes
}
