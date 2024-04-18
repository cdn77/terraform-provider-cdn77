package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.ConfigValidator = &OriginTypeConfigValidator{}
	_ resource.ConfigValidator   = &OriginTypeConfigValidator{}
)

type OriginTypeConfigValidator struct{}

func NewOriginTypeConfigValidator() *OriginTypeConfigValidator {
	return &OriginTypeConfigValidator{}
}

func (v OriginTypeConfigValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (OriginTypeConfigValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf(
		`Origin type "%s" can't be combined with attributes `+
			`"aws_access_key_id", "aws_access_key_secret" and "aws_region"`,
		OriginTypeUrl,
	)
}

func (v OriginTypeConfigValidator) ValidateDataSource(
	ctx context.Context,
	req datasource.ValidateConfigRequest,
	resp *datasource.ValidateConfigResponse,
) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v OriginTypeConfigValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (OriginTypeConfigValidator) Validate(ctx context.Context, config tfsdk.Config) diag.Diagnostics {
	var data OriginModel
	diags := config.Get(ctx, &data)

	switch data.Type.ValueString() {
	case OriginTypeUrl:
	default:
		return diags
	}

	conflictingAttributes := []struct {
		name  string
		value types.String
	}{
		{name: "aws_access_key_id", value: data.AwsAccessKeyId},
		{name: "aws_access_key_secret", value: data.AwsAccessKeySecret},
		{name: "aws_region", value: data.AwsRegion},
	}

	for _, attribute := range conflictingAttributes {
		if attribute.value.IsNull() {
			continue
		}

		diags.Append(
			validatordiag.InvalidAttributeCombinationDiagnostic(
				path.Root(attribute.name),
				fmt.Sprintf(`Attribute "%s" can't be used with Origin type "%s"'`, attribute.name, OriginTypeUrl),
			),
		)
	}

	return diags
}
