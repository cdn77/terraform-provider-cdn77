package provider

import (
	"context"
	"sort"
	"time"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

type CdnsModel struct {
	Cdns []CdnSummaryModel `tfsdk:"cdns"`
}

type CdnSummaryModel struct {
	Id                        types.Int64  `tfsdk:"id"`
	Cnames                    types.Set    `tfsdk:"cnames"`
	CreationTime              types.String `tfsdk:"creation_time"`
	Label                     types.String `tfsdk:"label"`
	Note                      types.String `tfsdk:"note"`
	OriginId                  types.String `tfsdk:"origin_id"`
	Url                       types.String `tfsdk:"url"`
	Mp4PseudoStreamingEnabled types.Bool   `tfsdk:"mp4_pseudo_streaming_enabled"`
}

var _ datasource.DataSourceWithConfigure = &CdnsDataSource{}

func NewCdnsDataSource() datasource.DataSource {
	return &CdnsDataSource{}
}

type CdnsDataSource struct {
	client cdn77.ClientWithResponsesInterface
}

func (*CdnsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cdns"
}

func (*CdnsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	cdnsSchema := util.NewResourceDataSourceSchemaConverter().Convert(CreateCdnResourceSchema())

	maps.DeleteFunc(cdnsSchema.Attributes, func(name string, _ schema.Attribute) bool {
		switch name {
		case "id", "cnames", "creation_time", "label", "mp4_pseudo_streaming_enabled", "note", "origin_id", "url":
			return false
		default:
			return true
		}
	})

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cdns": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{Attributes: cdnsSchema.Attributes},
				Computed:     true,
				Description:  "List of all CDNs",
			},
		},
		Description: "CDNs data source allows you to read all your CDNs",
	}
}

func (d *CdnsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.Append(util.MaybeSetClient(req.ProviderData, &d.client))
}

func (d *CdnsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	const errMessage = "Failed to fetch list of all CDNs"

	response, err := d.client.CdnListWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError(errMessage, err.Error())

		return
	}

	if !util.CheckResponse(&resp.Diagnostics, errMessage, response, response.JSONDefault) {
		return
	}

	cdns := make([]CdnSummaryModel, 0, len(*response.JSON200))

	for _, cdn := range *response.JSON200 {
		cnamesRaw := make([]string, len(cdn.Cnames))

		for i, c := range cdn.Cnames {
			cnamesRaw[i] = c.Cname
		}

		cnames, ds := types.SetValueFrom(ctx, types.StringType, cnamesRaw)
		if ds != nil {
			resp.Diagnostics.Append(ds...)

			return
		}

		mp4PseudoStreamingEnabled := false
		if cdn.Mp4PseudoStreaming != nil && cdn.Mp4PseudoStreaming.Enabled != nil {
			mp4PseudoStreamingEnabled = *cdn.Mp4PseudoStreaming.Enabled
		}

		cdns = append(cdns, CdnSummaryModel{
			Id:                        types.Int64Value(int64(cdn.Id)),
			Cnames:                    cnames,
			CreationTime:              types.StringValue(cdn.CreationTime.Format(time.DateTime)),
			Label:                     types.StringValue(cdn.Label),
			Note:                      util.NullableToStringValue(cdn.Note),
			OriginId:                  util.NullableToStringValue(cdn.OriginId),
			Url:                       types.StringValue(cdn.Url),
			Mp4PseudoStreamingEnabled: types.BoolValue(mp4PseudoStreamingEnabled),
		})
	}

	sort.SliceStable(cdns, func(i, j int) bool {
		return cdns[i].Id.ValueInt64() < cdns[j].Id.ValueInt64()
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &CdnsModel{Cdns: cdns})...)
}
