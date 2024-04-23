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

type SslsModel struct {
	Ssls []SslModel `tfsdk:"ssls"`
}

var _ datasource.DataSourceWithConfigure = &SslsDataSource{}

func NewSslsDataSource() datasource.DataSource {
	return &SslsDataSource{}
}

type SslsDataSource struct {
	client cdn77.ClientWithResponsesInterface
}

func (*SslsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssls"
}

func (*SslsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	sslSchema := util.NewResourceDataSourceSchemaConverter().Convert(CreateSslResourceSchema())
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ssls": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{Attributes: sslSchema.Attributes},
				Computed:     true,
				Description:  "List of all SSLs",
			},
		},
		Description: "SSLs data source allows you to read all your SSL certificates and keys",
	}
}

func (d *SslsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.Append(util.MaybeSetClient(req.ProviderData, &d.client))
}

func (d *SslsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	const errMessage = "Failed to fetch list of all SSLs"

	response, err := d.client.SslSniListWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError(errMessage, err.Error())

		return
	}

	if !util.CheckResponse(&resp.Diagnostics, errMessage, response, response.JSONDefault) {
		return
	}

	ssls := make([]SslModel, 0, len(*response.JSON200))

	for _, ssl := range *response.JSON200 {
		data := SslModel{Id: types.StringValue(ssl.Id)}
		if ds := readSslDetails(ctx, &data, &ssl); ds != nil { //nolint:gosec // false positive since go1.22
			resp.Diagnostics.Append(ds...)

			return
		}

		ssls = append(ssls, data)
	}

	sort.SliceStable(ssls, func(i, j int) bool {
		return ssls[i].Id.ValueString() < ssls[j].Id.ValueString()
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &SslsModel{Ssls: ssls})...)
}
