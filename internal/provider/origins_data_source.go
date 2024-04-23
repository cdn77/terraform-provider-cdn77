package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OriginsModel struct {
	Origins []OriginModel `tfsdk:"origins"`
}

var _ datasource.DataSourceWithConfigure = &OriginsDataSource{}

func NewOriginsDataSource() datasource.DataSource {
	return &OriginsDataSource{}
}

type OriginsDataSource struct {
	client cdn77.ClientWithResponsesInterface
}

func (*OriginsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_origins"
}

func (*OriginsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	originSchema := util.NewResourceDataSourceSchemaConverter().Convert(CreateOriginResourceSchema())
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"origins": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{Attributes: originSchema.Attributes},
				Computed:     true,
				Description:  "List of all Origins",
			},
		},
		Description: "Origins data source allows you to read all your Origins",
	}
}

func (d *OriginsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.Append(util.MaybeSetClient(req.ProviderData, &d.client))
}

func (d *OriginsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	const errMessage = "Failed to fetch list of all Origins"

	response, err := d.client.OriginListWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError(errMessage, err.Error())

		return
	}

	if !util.CheckResponse(&resp.Diagnostics, errMessage, response, response.JSONDefault) {
		return
	}

	origins := make([]OriginModel, 0, len(*response.JSON200))

	for _, originUnion := range *response.JSON200 {
		origin, err := originUnion.ValueByDiscriminator()
		if err != nil {
			resp.Diagnostics.AddError(errMessage, fmt.Sprintf("Failed to convert Origin to a specific type: %s", err))

			return
		}

		var originModel OriginModel

		switch o := origin.(type) {
		case cdn77.S3OriginDetail:
			originModel = OriginModel{
				Id:             types.StringValue(o.Id),
				Type:           types.StringValue(OriginTypeAws),
				Label:          types.StringValue(o.Label),
				Note:           util.NullableToStringValue(o.Note),
				AwsAccessKeyId: util.NullableToStringValue(o.AwsAccessKeyId),
				AwsRegion:      util.NullableToStringValue(o.AwsRegion),
				Scheme:         types.StringValue(string(o.Scheme)),
				Host:           types.StringValue(o.Host),
				Port:           util.NullableIntToInt64Value(o.Port),
				BaseDir:        util.NullableToStringValue(o.BaseDir),
			}
		case cdn77.ObjectStorageOriginDetail:
			originModel = OriginModel{
				Id:         types.StringValue(o.Id),
				Type:       types.StringValue(OriginTypeObjectStorage),
				Label:      types.StringValue(o.Label),
				Note:       util.NullableToStringValue(o.Note),
				BucketName: types.StringValue(o.BucketName),
				Scheme:     types.StringValue(string(o.Scheme)),
				Host:       types.StringValue(o.Host),
				Port:       util.NullableIntToInt64Value(o.Port),
			}
		case cdn77.UrlOriginDetail:
			originModel = OriginModel{
				Id:      types.StringValue(o.Id),
				Type:    types.StringValue(OriginTypeUrl),
				Label:   types.StringValue(o.Label),
				Note:    util.NullableToStringValue(o.Note),
				Scheme:  types.StringValue(string(o.Scheme)),
				Host:    types.StringValue(o.Host),
				Port:    util.NullableIntToInt64Value(o.Port),
				BaseDir: util.NullableToStringValue(o.BaseDir),
			}
		default:
			continue
		}

		origins = append(origins, originModel)
	}

	sort.SliceStable(origins, func(i, j int) bool {
		return origins[i].Id.ValueString() < origins[j].Id.ValueString()
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &OriginsModel{Origins: origins})...)
}
