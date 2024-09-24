package origin

import (
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsModel struct {
	AwsBaseModel

	AccessKeySecret types.String `tfsdk:"access_key_secret"`
}

type AwsBaseModel struct {
	SharedModel
	shared.UrlModel

	AccessKeyId types.String `tfsdk:"access_key_id"`
	Region      types.String `tfsdk:"region"`
}

func CreateAwsResourceSchema() schema.Schema {
	s := CreateAwsBaseResourceSchema()
	s.Attributes["access_key_secret"] = schema.StringAttribute{
		Description: "AWS access key secret",
		Optional:    true,
		Sensitive:   true,
		Validators: []validator.String{
			stringvalidator.AlsoRequires(path.MatchRoot("access_key_id"), path.MatchRoot("region")),
		},
	}

	return s
}

func CreateAwsBaseResourceSchema() schema.Schema {
	return WithSharedSchemaAttrs(shared.WithUrlSchemaAttrs(schema.Schema{
		MarkdownDescription: "AWS Origin resource allows you to manage your AWS Origins",
		Attributes: map[string]schema.Attribute{
			"access_key_id": schema.StringAttribute{
				Description: "AWS access key ID",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("access_key_secret"), path.MatchRoot("region")),
				},
			},
			"region": schema.StringAttribute{
				Description: "AWS region",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("access_key_secret"), path.MatchRoot("access_key_id")),
				},
			},
		},
	}))
}
