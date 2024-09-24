package object_storages

import (
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/shared"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	ds_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rsc_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AllModel struct {
	Clusters []Model `tfsdk:"clusters"`
}

type Model struct {
	shared.UrlModel

	Id    types.String `tfsdk:"id"`
	Label types.String `tfsdk:"label"`
}

func CreateSchema() ds_schema.Schema {
	s := util.NewResourceDataSourceSchemaConverter().Convert(shared.WithComputedUrlSchemaAttrs(rsc_schema.Schema{
		Attributes: map[string]rsc_schema.Attribute{
			"id": rsc_schema.StringAttribute{
				Description: "ID (UUID) of the Object Storage cluster",
				Computed:    true,
			},
			"label": rsc_schema.StringAttribute{
				Computed:    true,
				Description: "Label of the Object Storage cluster",
			},
		},
	}))

	return ds_schema.Schema{
		Attributes: map[string]ds_schema.Attribute{
			"clusters": ds_schema.ListNestedAttribute{
				NestedObject: ds_schema.NestedAttributeObject{Attributes: s.Attributes},
				Computed:     true,
				Description:  "List of all Object Storage clusters",
			},
		},
		Description: "Object Storages data source allows you to read all available Object Storage clusters",
	}
}
