package ssl

import (
	"cmp"
	"context"
	"slices"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AllModel struct {
	Ssls []BaseModel `tfsdk:"ssls"`
}

var _ datasource.DataSourceWithConfigure = &AllDataSource{}

type AllDataSource struct {
	*util.BaseDataSource
}

func (d *AllDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	d.BaseDataSource.Schema(ctx, req, resp)

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ssls": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{Attributes: resp.Schema.Attributes},
				Computed:     true,
				Description:  "List of all SSLs",
			},
		},
		Description: "SSLs data source allows you to read all your SSL certificates and keys",
	}
}

func (d *AllDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	const errMessage = "Failed to fetch list of all SSLs"

	diags := &resp.Diagnostics

	response, err := d.Client.SslSniListWithResponse(ctx)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ProcessResponse(diags, response, errMessage, response.JSON200, func(list *cdn77.SslList) {
		ssls := make([]BaseModel, 0, len(*list))

		for _, ssl := range *list {
			data := BaseModel{Id: types.StringValue(ssl.Id)}
			if readSslDetails(ctx, diags, &data, &ssl); diags.HasError() {
				return
			}

			ssls = append(ssls, data)
		}

		slices.SortStableFunc(ssls, func(a, b BaseModel) int {
			return cmp.Compare(a.Id.ValueString(), b.Id.ValueString())
		})
		diags.Append(resp.State.Set(ctx, AllModel{Ssls: ssls})...)
	})
}
