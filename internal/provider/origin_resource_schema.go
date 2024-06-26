package provider

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	OriginTypeAws           = "aws"
	OriginTypeObjectStorage = "object-storage"
	OriginTypeUrl           = "url"
)

var originTypes = []string{OriginTypeAws, OriginTypeObjectStorage, OriginTypeUrl} //nolint:gochecknoglobals

func CreateOriginResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Origin resource allows you to manage your Origins",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Origin ID (UUID)",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"type": schema.StringAttribute{
				Description:   fmt.Sprintf("Type of the origin; one of %v", originTypes),
				Required:      true,
				Validators:    []validator.String{stringvalidator.OneOf(originTypes...)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"label": schema.StringAttribute{
				Description: "The label helps you to identify your Origin",
				Required:    true,
			},
			"note": schema.StringAttribute{
				Description: "Optional note for the Origin",
				Optional:    true,
			},
			"aws_access_key_id": schema.StringAttribute{
				Description: "AWS access key ID",
				Optional:    true,
			},
			"aws_access_key_secret": schema.StringAttribute{
				Description: "AWS access key secret",
				Optional:    true,
				Sensitive:   true,
			},
			"aws_region": schema.StringAttribute{
				Description: "AWS region",
				Optional:    true,
			},
			"acl": schema.StringAttribute{
				Description: "Object Storage access key ACL",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("authenticated-read", "private", "public-read", "public-read-write"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"cluster_id": schema.StringAttribute{
				Description:   "ID of the Object Storage storage cluster",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"bucket_name": schema.StringAttribute{
				Description: "Name of your Object Storage bucket",
				Optional:    true,
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
			"access_key_id": schema.StringAttribute{
				Description:   "Access key to your Object Storage bucket",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"access_key_secret": schema.StringAttribute{
				Description:   "Access secret to your Object Storage bucket",
				Computed:      true,
				Sensitive:     true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"scheme": schema.StringAttribute{
				Description: "Scheme of the Origin",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.String{stringvalidator.OneOf("http", "https")},
			},
			"host": schema.StringAttribute{
				Description: "Origin host without scheme and port. Can be domain name or IP address",
				Optional:    true,
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Origin port number. If not specified, default scheme port is used",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.Int64{int64validator.Between(1, 65535)},
			},
			"base_dir": schema.StringAttribute{
				Description: "Directory where the content is stored on the URL Origin",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
					stringvalidator.RegexMatches(regexp.MustCompile(`[^/]$`), "base_dir mustn't end with the slash"),
				},
			},
		},
	}
}
