package provider //nolint:dupl // false-positive

import (
	"context"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSourceWithConfigure = &SslDataSource{}

func NewSslDataSource() datasource.DataSource {
	return &SslDataSource{}
}

type SslDataSource struct {
	client cdn77.ClientWithResponsesInterface
}

func (*SslDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssl"
}

func (*SslDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = util.NewResourceDataSourceSchemaConverter("id").Convert(CreateSslResourceSchema())
}

func (d *SslDataSource) Configure(_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.Append(util.MaybeSetClient(req.ProviderData, &d.client))
}

func (d *SslDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	NewSslDataSourceReader(ctx, d.client).Read(&req.Config, &resp.Diagnostics, &resp.State)
}
