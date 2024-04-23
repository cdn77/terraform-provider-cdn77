package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	_ datasource.ConfigValidator = &OriginConfigValidator{}
	_ resource.ConfigValidator   = &OriginConfigValidator{}
)

type OriginConfigValidator struct{}

func NewOriginTypeConfigValidator() *OriginConfigValidator {
	return &OriginConfigValidator{}
}

func (v OriginConfigValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (OriginConfigValidator) MarkdownDescription(_ context.Context) string {
	return "Checks Origin configuration for all required/conflicting attributes"
}

func (v OriginConfigValidator) ValidateDataSource(
	ctx context.Context,
	req datasource.ValidateConfigRequest,
	resp *datasource.ValidateConfigResponse,
) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v OriginConfigValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v OriginConfigValidator) Validate(ctx context.Context, config tfsdk.Config) diag.Diagnostics {
	var data OriginModel
	diags := config.Get(ctx, &data)

	originType := data.Type.ValueString()
	awsAttributes := []attrNameAndValue{
		{name: "aws_access_key_id", value: data.AwsAccessKeyId},
		{name: "aws_access_key_secret", value: data.AwsAccessKeySecret},
		{name: "aws_region", value: data.AwsRegion},
	}
	objectStorageAttributes := []attrNameAndValue{
		{name: "acl", value: data.Acl},
		{name: "cluster_id", value: data.ClusterId},
		{name: "bucket_name", value: data.BucketName},
	}
	schemeAndHostAttributes := []attrNameAndValue{
		{name: "scheme", value: data.Scheme},
		{name: "host", value: data.Host},
	}
	var conflictingAttributes, requiredAttributes []attrNameAndValue

	switch originType {
	case OriginTypeAws:
		conflictingAttributes = append(conflictingAttributes, objectStorageAttributes...)
		requiredAttributes = schemeAndHostAttributes
	case OriginTypeObjectStorage:
		conflictingAttributes = append(conflictingAttributes, awsAttributes...)
		conflictingAttributes = append(conflictingAttributes, schemeAndHostAttributes...)
		conflictingAttributes = append(conflictingAttributes, attrNameAndValue{name: "port", value: data.Port})
		conflictingAttributes = append(conflictingAttributes, attrNameAndValue{name: "base_dir", value: data.BaseDir})
		requiredAttributes = objectStorageAttributes
	case OriginTypeUrl:
		conflictingAttributes = append(conflictingAttributes, awsAttributes...)
		conflictingAttributes = append(conflictingAttributes, objectStorageAttributes...)
		requiredAttributes = schemeAndHostAttributes
	default:
		addUnknownOriginTypeError(&diags, data)

		return diags
	}

	diags.Append(v.doValidate(originType, conflictingAttributes, requiredAttributes)...)

	return diags
}

func (OriginConfigValidator) doValidate(
	originType string,
	conflictingAttributes []attrNameAndValue,
	requiredAttributes []attrNameAndValue,
) (diags diag.Diagnostics) {
	for _, attribute := range conflictingAttributes {
		if attribute.value.IsNull() {
			continue
		}

		diags.Append(
			validatordiag.InvalidAttributeCombinationDiagnostic(
				path.Root(attribute.name),
				fmt.Sprintf(`Attribute "%s" can't be used with Origin type "%s"'`, attribute.name, originType),
			),
		)
	}

	for _, attribute := range requiredAttributes {
		if !attribute.value.IsNull() {
			continue
		}

		diags.Append(
			validatordiag.InvalidAttributeCombinationDiagnostic(
				path.Root(attribute.name),
				fmt.Sprintf(`Attribute "%s" is required for Origin type "%s"'`, attribute.name, originType),
			),
		)
	}

	return diags
}

type attrNameAndValue struct {
	name  string
	value attr.Value
}
