package ssl

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
	BaseModel

	PrivateKey types.String `tfsdk:"private_key"`
}

type BaseModel struct {
	Id          types.String `tfsdk:"id"`
	Certificate types.String `tfsdk:"certificate"`
	Subjects    types.Set    `tfsdk:"subjects"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
}

func CreateResourceSchema() schema.Schema {
	s := CreateBaseResourceSchema()
	s.Attributes["private_key"] = schema.StringAttribute{
		Description: "Private key associated with the certificate",
		Required:    true,
		Sensitive:   true,
	}

	return s
}

func CreateBaseResourceSchema() schema.Schema {
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
