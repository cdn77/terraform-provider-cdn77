package provider

import (
	"context"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSourceWithConfigure = &OriginDataSource{}

func NewOriginDataSource() datasource.DataSource {
	return &OriginDataSource{}
}

type OriginDataSource struct {
	client cdn77.ClientWithResponsesInterface
}

func (*OriginDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_origin"
}

func (*OriginDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = util.NewResourceDataSourceSchemaConverter("id", "type").Convert(CreateOriginResourceSchema())
}

func (d *OriginDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.Append(util.MaybeSetClient(req.ProviderData, &d.client))
}

func (d *OriginDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	NewOriginDataSourceReader(ctx, d.client).Read(&req.Config, &resp.Diagnostics, &resp.State)
}
