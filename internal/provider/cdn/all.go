package cdn

import (
	"cmp"
	"context"
	"slices"
	"sync"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AllModel struct {
	Cdns []Model `tfsdk:"cdns"`
}

var _ datasource.DataSourceWithConfigure = &AllDataSource{}

type AllDataSource struct {
	*util.BaseDataSource
}

func (d *AllDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	d.BaseDataSource.Schema(ctx, req, resp)

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cdns": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{Attributes: resp.Schema.Attributes},
				Computed:     true,
				Description:  "List of all CDNs",
			},
		},
		Description: "CDNs data source allows you to read all your CDNs",
	}
}

func (d *AllDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	const errMessage = "Failed to fetch list of all CDNs"

	diags := &resp.Diagnostics

	response, err := d.Client.CdnListWithResponse(ctx)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ProcessResponse(diags, response, errMessage, response.JSON200, func(summaries *[]cdn77.CdnSummary) {
		wg := sync.WaitGroup{}
		mu := sync.Mutex{}
		cdns := make([]Model, 0, len(*summaries))

		wg.Add(len(*summaries))

		for _, summary := range *summaries {
			go func() {
				var data any = Model{Id: types.Int64Value(int64(summary.Id))}

				ds := d.BaseDataSource.Reader().Fill(ctx, d.Client, &data)

				mu.Lock()

				if ds == nil {
					cdns = append(cdns, data.(Model))
				} else {
					diags.Append(ds...)
				}

				mu.Unlock()
				wg.Done()
			}()
		}

		if wg.Wait(); !diags.HasError() {
			slices.SortStableFunc(cdns, func(a, b Model) int {
				return cmp.Compare(a.Id.ValueInt64(), b.Id.ValueInt64())
			})
			diags.Append(resp.State.Set(ctx, AllModel{Cdns: cdns})...)
		}
	})
}
