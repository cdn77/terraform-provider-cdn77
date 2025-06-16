package origin

import (
	"regexp"

	"github.com/cdn77/terraform-provider-cdn77/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ObjectStorageModel struct {
	ObjectStorageBaseModel

	Acl       types.String `tfsdk:"acl"`
	ClusterId types.String `tfsdk:"cluster_id"`
}

type ObjectStorageBaseModel struct {
	SharedModel
	shared.UrlModel

	BucketName types.String             `tfsdk:"bucket_name"`
	Usage      *ObjectStorageUsageModel `tfsdk:"usage"`
}

type ObjectStorageUsageModel struct {
	Files     types.Int64 `tfsdk:"files"`
	SizeBytes types.Int64 `tfsdk:"size_bytes"`
}

func CreateObjectStorageResourceSchema() schema.Schema {
	s := CreateObjectStorageBaseResourceSchema()
	s.Attributes["acl"] = schema.StringAttribute{
		Description: "Object Storage access key ACL",
		Required:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("authenticated-read", "private", "public-read", "public-read-write"),
		},
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
	s.Attributes["cluster_id"] = schema.StringAttribute{
		Description:   "ID of the Object Storage storage cluster",
		Required:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}

	return s
}

func CreateObjectStorageBaseResourceSchema() schema.Schema {
	return WithSharedSchemaAttrs(shared.WithComputedUrlSchemaAttrs(schema.Schema{
		MarkdownDescription: "Object Storage Origin resource allows you to manage your Object Storage Origins",
		Attributes: map[string]schema.Attribute{
			"bucket_name": schema.StringAttribute{
				Description: "Name of your Object Storage bucket",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^([a-z0-9][a-z0-9-]{1,61}[a-z0-9])?$`),
						"Allowed characters are lowercase letters, digits and a dash. "+
							"Dash isn't allowed at the start and end of the bucket name.",
					),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"usage": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"files": schema.Int64Attribute{
						Computed:      true,
						Description:   "Number of files stored on the Object Storage bucket",
						PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					},
					"size_bytes": schema.Int64Attribute{
						Computed:      true,
						Description:   "Total size of the Object Storage bucket in bytes",
						PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					},
				},
				Computed:    true,
				Description: "Usage statistics of the Object Storage bucket",
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{"files": types.Int64Type, "size_bytes": types.Int64Type},
					map[string]attr.Value{
						"files":      types.Int64Value(0),
						"size_bytes": types.Int64Value(0),
					},
				)),
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			},
		},
	}))
}
