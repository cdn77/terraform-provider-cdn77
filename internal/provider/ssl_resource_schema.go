package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func CreateSslResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "SSL resource allows you to managed your SSL certificates and keys",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "ID (UUID) of the SSL certificate",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"certificate": schema.StringAttribute{
				Description: "SNI certificate",
				Required:    true,
			},
			"private_key": schema.StringAttribute{
				Description: "Private key associated with the certificate",
				Required:    true,
				Sensitive:   true,
			},
			"subjects": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Subjects (domain names) of the certificate",
			},
			"expires_at": schema.StringAttribute{
				Description: "Date and time of the SNI certificate expiration",
				Computed:    true,
			},
		},
	}
}
