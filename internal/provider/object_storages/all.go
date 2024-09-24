package object_storages

import (
	"context"
	"sort"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/shared"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oapi-codegen/nullable"
)

var _ datasource.DataSourceWithConfigure = &DataSource{}

type DataSource struct {
	*util.BaseDataSource
}

func (d *DataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	const errMessage = "Failed to fetch list of all Object Storages"

	diags := &resp.Diagnostics

	response, err := d.Client.ObjectStorageClusterListWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError(errMessage, err.Error())

		return
	}

	util.ProcessResponse(diags, response, errMessage, response.JSON200, func(list *cdn77.ObjectStorageClusters) {
		objectStorages := make([]Model, 0, len(*list))

		for _, os := range *list {
			port := nullable.NewNullNullable[int]()
			if os.Port != nil {
				port.Set(*os.Port)
			}

			objectStorages = append(objectStorages, Model{
				UrlModel: shared.NewUrlModel(ctx, os.Scheme, os.Host, port, nullable.NewNullNullable[string]()),
				Id:       types.StringValue(os.Id),
				Label:    types.StringValue(os.Label),
			})
		}

		sort.SliceStable(objectStorages, func(i, j int) bool {
			return objectStorages[i].Id.ValueString() < objectStorages[j].Id.ValueString()
		})

		resp.Diagnostics.Append(resp.State.Set(ctx, AllModel{Clusters: objectStorages})...)
	})
}
