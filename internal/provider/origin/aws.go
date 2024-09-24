package origin

import (
	"context"
	"fmt"
	"strings"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/shared"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &AwsResource{}
	_ resource.ResourceWithImportState = &AwsResource{}
	_ resource.ResourceWithMoveState   = &AwsResource{}
)

type AwsResource struct {
	*util.BaseResource
}

func (r *AwsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	diags := &resp.Diagnostics
	var data AwsModel

	if diags.Append(req.Plan.Get(ctx, &data)...); diags.HasError() {
		return
	}

	const errMessage = "Failed to create AWS Origin"

	scheme, host, port, basePath := data.UrlModel.Parts(ctx)
	request := cdn77.OriginCreateAwsJSONRequestBody{
		Label:              data.Label.ValueString(),
		Note:               util.StringValueToNullable(data.Note),
		Scheme:             cdn77.OriginScheme(scheme),
		Host:               host,
		Port:               port,
		BaseDir:            basePath,
		AwsAccessKeyId:     util.StringValueToNullable(data.AccessKeyId),
		AwsAccessKeySecret: util.StringValueToNullable(data.AccessKeySecret),
		AwsRegion:          util.StringValueToNullable(data.Region),
	}

	response, err := r.Client.OriginCreateAwsWithResponse(ctx, request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ProcessResponse(diags, response, errMessage, response.JSON201, func(detail *cdn77.S3OriginDetail) {
		data.Id = types.StringValue(detail.Id)

		diags.Append(resp.State.Set(ctx, data)...)
	})
}

func (r *AwsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := &resp.Diagnostics
	var data AwsModel

	if diags.Append(req.Plan.Get(ctx, &data)...); diags.HasError() {
		return
	}

	const errMessage = "Failed to update AWS Origin"

	scheme, host, port, basePath := data.UrlModel.Parts(ctx)
	request := cdn77.OriginEditAwsJSONRequestBody{
		Label:              data.Label.ValueStringPointer(),
		Note:               util.StringValueToNullable(data.Note),
		Scheme:             util.Pointer(cdn77.OriginScheme(scheme)),
		Host:               &host,
		Port:               port,
		BaseDir:            basePath,
		AwsAccessKeyId:     util.StringValueToNullable(data.AccessKeyId),
		AwsAccessKeySecret: util.StringValueToNullable(data.AccessKeySecret),
		AwsRegion:          util.StringValueToNullable(data.Region),
	}

	response, err := r.Client.OriginEditAwsWithResponse(ctx, data.Id.ValueString(), request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ProcessEmptyResponse(diags, response, errMessage, func() {
		diags.Append(resp.State.Set(ctx, data)...)
	})
}

func (r *AwsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	diags := &resp.Diagnostics
	var data AwsModel

	if diags.Append(req.State.Get(ctx, &data)...); diags.HasError() {
		return
	}

	const errMessage = "Failed to delete AWS Origin"

	response, err := r.Client.OriginDeleteAwsWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ValidateDeletionResponse(diags, response, errMessage)
}

func (*AwsResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) < 2 {
		resp.Diagnostics.AddError(
			"Invalid Import Identifier",
			fmt.Sprintf("Expected: <id>,<access_key_secret>\nGot: %q", req.ID),
		)

		return
	}

	id, accessKeySecret := idParts[0], idParts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)

	if accessKeySecret != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("access_key_secret"), accessKeySecret)...)
	}
}

func (r *AwsResource) MoveState(context.Context) []resource.StateMover {
	deprecatedOriginSchema := CreateDeprecatedOriginResourceSchema()

	return []resource.StateMover{
		{
			SourceSchema: &deprecatedOriginSchema,
			StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
				if req.SourceTypeName != DeprecatedOriginResourceType ||
					!strings.HasSuffix(req.SourceProviderAddress, "cdn77/cdn77") {
					return
				}

				diags := &resp.Diagnostics
				var oldModel DeprecatedOriginModel

				if diags.Append(req.SourceState.Get(ctx, &oldModel)...); diags.HasError() {
					return
				}

				if oldModel.Type.ValueString() != TypeAws {
					diags.AddError(
						"Unable to Move Resource State",
						fmt.Sprintf(
							"Only %q resources with type=%q can be moved to %q.\nLabel: %s\nType: %s\n",
							DeprecatedOriginResourceType,
							TypeAws,
							r.FullName(),
							oldModel.Label.ValueString(),
							oldModel.Type.ValueString(),
						),
					)

					return
				}

				model := AwsModel{
					AwsBaseModel: AwsBaseModel{
						SharedModel: NewSharedModel(
							oldModel.Id,
							oldModel.Label.ValueString(),
							util.StringValueToNullable(oldModel.Note),
						),
						UrlModel: shared.NewUrlModel(
							ctx,
							oldModel.Scheme.ValueString(),
							oldModel.Host.ValueString(),
							util.Int64ValueToNullable[int](oldModel.Port),
							util.StringValueToNullable(oldModel.BaseDir),
						),
						AccessKeyId: oldModel.AwsAccessKeyId,
						Region:      oldModel.AwsRegion,
					},
					AccessKeySecret: oldModel.AwsAccessKeySecret,
				}

				diags.Append(resp.TargetState.Set(ctx, model)...)
			},
		},
	}
}

var _ datasource.DataSourceWithConfigure = &AwsDataSource{}

type AwsDataSource struct {
	*util.BaseDataSource
}
