package util

import (
	"context"
	"strings"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	ds_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsc_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type BaseResource struct {
	name           string
	schemaProvider func() rsc_schema.Schema
	reader         Reader

	providerTypeName string
	Client           cdn77.ClientWithResponsesInterface
}

func NewBaseResource(name string, schemaProvider func() rsc_schema.Schema, reader Reader) *BaseResource {
	return &BaseResource{name: name, schemaProvider: schemaProvider, reader: reader}
}

func (r *BaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	r.providerTypeName = req.ProviderTypeName
	resp.TypeName = strings.Join([]string{r.providerTypeName, r.name}, "_")
}

func (r *BaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.schemaProvider()
}

func (r *BaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(MaybeSetClient(req.ProviderData, &r.Client))
}

func (r *BaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.reader.RemoveMissingResource().Read(ctx, r.Client, &req.State, &resp.State, &resp.Diagnostics)
}

func (r *BaseResource) FullName() string {
	return r.providerTypeName + "_" + r.name
}

type BaseDataSource struct {
	name           string
	schemaProvider func() ds_schema.Schema
	reader         Reader

	Client cdn77.ClientWithResponsesInterface
}

func NewBaseDataSource(name string, schemaProvider func() ds_schema.Schema, reader Reader) *BaseDataSource {
	return &BaseDataSource{name: name, schemaProvider: schemaProvider, reader: reader}
}

func (d *BaseDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = strings.Join([]string{req.ProviderTypeName, d.name}, "_")
}

func (d *BaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = d.schemaProvider()
}

func (d *BaseDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.Append(MaybeSetClient(req.ProviderData, &d.Client))
}

func (d *BaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	d.reader.Read(ctx, d.Client, &req.Config, &resp.State, &resp.Diagnostics)
}

func (d *BaseDataSource) Reader() Reader {
	return d.reader
}
