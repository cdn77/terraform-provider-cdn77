package provider

import (
	"context"
	"sort"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ObjectStoragesModel struct {
	Clusters []ObjectStorageModel `tfsdk:"clusters"`
}

type ObjectStorageModel struct {
	Id     types.String `tfsdk:"id"`
	Host   types.String `tfsdk:"host"`
	Label  types.String `tfsdk:"label"`
	Port   types.Int64  `tfsdk:"port"`
	Scheme types.String `tfsdk:"scheme"`
}

var _ datasource.DataSourceWithConfigure = &ObjectStoragesDataSource{}

func NewObjectStoragesDataSource() datasource.DataSource {
	return &ObjectStoragesDataSource{}
}

type ObjectStoragesDataSource struct {
	client cdn77.ClientWithResponsesInterface
}

func (*ObjectStoragesDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_object_storages"
}

func (*ObjectStoragesDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"clusters": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID (UUID) of the Object Storage cluster",
							Computed:    true,
						},
						"host": schema.StringAttribute{
							Computed: true,
						},
						"label": schema.StringAttribute{
							Computed: true,
						},
						"port": schema.Int64Attribute{
							Computed: true,
						},
						"scheme": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed:    true,
				Description: "List of all Object Storage clusters",
			},
		},
		Description: "Object Storages data source allows you to read all available Object Storage clusters",
	}
}

func (d *ObjectStoragesDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.Append(util.MaybeSetClient(req.ProviderData, &d.client))
}

func (d *ObjectStoragesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	const errMessage = "Failed to fetch list of all Object Storages"

	response, err := d.client.ObjectStorageClusterListWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError(errMessage, err.Error())

		return
	}

	if !util.CheckResponse(&resp.Diagnostics, errMessage, response, response.JSONDefault) {
		return
	}

	objectStorages := make([]ObjectStorageModel, 0, len(*response.JSON200))

	for _, objectStorage := range *response.JSON200 {
		objectStorages = append(objectStorages, ObjectStorageModel{
			Id:     types.StringValue(objectStorage.Id),
			Host:   types.StringValue(objectStorage.Host),
			Label:  types.StringValue(objectStorage.Label),
			Port:   util.IntPointerToInt64Value(objectStorage.Port),
			Scheme: types.StringValue(objectStorage.Scheme),
		})
	}

	sort.SliceStable(objectStorages, func(i, j int) bool {
		return objectStorages[i].Id.ValueString() < objectStorages[j].Id.ValueString()
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &ObjectStoragesModel{Clusters: objectStorages})...)
}
