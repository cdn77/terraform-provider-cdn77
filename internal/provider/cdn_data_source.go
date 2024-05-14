package provider //nolint:dupl // false-positive

import (
	"context"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSourceWithConfigure = &CdnDataSource{}

func NewCdnDataSource() datasource.DataSource {
	return &CdnDataSource{}
}

type CdnDataSource struct {
	client cdn77.ClientWithResponsesInterface
}

func (*CdnDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cdn"
}

func (*CdnDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = util.NewResourceDataSourceSchemaConverter("id").Convert(CreateCdnResourceSchema())
}

func (d *CdnDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.Append(util.MaybeSetClient(req.ProviderData, &d.client))
}

func (d *CdnDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	NewCdnDataSourceReader(ctx, d.client).Read(&req.Config, &resp.Diagnostics, &resp.State)
}
