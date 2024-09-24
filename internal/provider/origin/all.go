package origin

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/shared"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oapi-codegen/nullable"
)

type AllModel struct {
	Aws           []AwsBaseModel           `tfsdk:"aws"`
	ObjectStorage []ObjectStorageBaseModel `tfsdk:"object_storage"`
	Url           []UrlModel               `tfsdk:"url"`
}

var _ datasource.DataSourceWithConfigure = &AllDataSource{}

type AllDataSource struct {
	client cdn77.ClientWithResponsesInterface
}

func NewAllDataSource() datasource.DataSource {
	return &AllDataSource{}
}

func (*AllDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = strings.Join([]string{req.ProviderTypeName, "origins"}, "_")
}

func (*AllDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	converter := util.NewResourceDataSourceSchemaConverter()
	awsAttrs := converter.Convert(CreateAwsBaseResourceSchema()).Attributes
	objectStorageAttrs := converter.Convert(CreateObjectStorageBaseResourceSchema()).Attributes
	urlAttrs := converter.Convert(CreateUrlResourceSchema()).Attributes

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aws": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{Attributes: awsAttrs},
				Computed:     true,
				Description:  "List of all AWS Origins",
			},
			"object_storage": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{Attributes: objectStorageAttrs},
				Computed:     true,
				Description:  "List of all Object Storage Origins",
			},
			"url": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{Attributes: urlAttrs},
				Computed:     true,
				Description:  "List of all URL Origins",
			},
		},
		Description: "Origins data source allows you to read all your Origins",
	}
}

func (d *AllDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.Append(util.MaybeSetClient(req.ProviderData, &d.client))
}

func (d *AllDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	const errMessage = "Failed to fetch list of all Origins"

	diags := &resp.Diagnostics

	response, err := d.client.OriginListWithResponse(ctx)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ProcessResponse(diags, response, errMessage, response.JSON200, func(list *cdn77.OriginList) {
		data := AllModel{Aws: []AwsBaseModel{}, ObjectStorage: []ObjectStorageBaseModel{}, Url: []UrlModel{}}

		for _, item := range *list {
			origin, err := item.ValueByDiscriminator()

			switch o := origin.(type) {
			case cdn77.S3OriginDetail:
				data.Aws = append(data.Aws, AwsBaseModel{
					SharedModel: NewSharedModel(types.StringValue(o.Id), o.Label, o.Note),
					UrlModel:    shared.NewUrlModel(ctx, string(o.Scheme), o.Host, o.Port, o.BaseDir),
					AccessKeyId: util.NullableToStringValue(o.AwsAccessKeyId),
					Region:      util.NullableToStringValue(o.AwsRegion),
				})
			case cdn77.ObjectStorageOriginDetail:
				data.ObjectStorage = append(data.ObjectStorage, ObjectStorageBaseModel{
					SharedModel: NewSharedModel(types.StringValue(o.Id), o.Label, o.Note),
					UrlModel: shared.NewUrlModel(
						ctx,
						string(o.Scheme),
						o.Host,
						o.Port,
						nullable.NewNullNullable[string](),
					),
					BucketName: types.StringValue(o.BucketName),
					Usage: &ObjectStorageUsageModel{
						Files:     util.IntPointerToInt64Value(o.Usage.FileCount),
						SizeBytes: util.IntPointerToInt64Value(o.Usage.SizeBytes),
					},
				})
			case cdn77.UrlOriginDetail:
				data.Url = append(data.Url, UrlModel{
					SharedModel: NewSharedModel(types.StringValue(o.Id), o.Label, o.Note),
					UrlModel:    shared.NewUrlModel(ctx, string(o.Scheme), o.Host, o.Port, o.BaseDir),
				})
			default:
				diags.AddError(errMessage, fmt.Sprintf("Failed to convert Origin to a specific type: %s", err))

				return
			}
		}

		slices.SortStableFunc(data.Aws, func(a, b AwsBaseModel) int {
			return cmp.Compare(a.Id.ValueString(), b.Id.ValueString())
		})
		slices.SortStableFunc(data.ObjectStorage, func(a, b ObjectStorageBaseModel) int {
			return cmp.Compare(a.Id.ValueString(), b.Id.ValueString())
		})
		slices.SortStableFunc(data.Url, func(a, b UrlModel) int {
			return cmp.Compare(a.Id.ValueString(), b.Id.ValueString())
		})
		diags.Append(resp.State.Set(ctx, data)...)
	})
}
