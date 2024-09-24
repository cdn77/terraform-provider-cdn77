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
	_ resource.ResourceWithConfigure   = &UrlResource{}
	_ resource.ResourceWithImportState = &UrlResource{}
	_ resource.ResourceWithMoveState   = &UrlResource{}
)

type UrlResource struct {
	*util.BaseResource
}

func (r *UrlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	diags := &resp.Diagnostics
	var data UrlModel

	if diags.Append(req.Plan.Get(ctx, &data)...); diags.HasError() {
		return
	}

	const errMessage = "Failed to create URL Origin"

	scheme, host, port, basePath := data.UrlModel.Parts(ctx)
	request := cdn77.OriginCreateUrlJSONRequestBody{
		Label:   data.Label.ValueString(),
		Note:    util.StringValueToNullable(data.Note),
		Scheme:  cdn77.OriginScheme(scheme),
		Host:    host,
		Port:    port,
		BaseDir: basePath,
	}

	response, err := r.Client.OriginCreateUrlWithResponse(ctx, request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ProcessResponse(diags, response, errMessage, response.JSON201, func(detail *cdn77.UrlOriginDetail) {
		data.Id = types.StringValue(detail.Id)

		diags.Append(resp.State.Set(ctx, data)...)
	})
}

func (r *UrlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := &resp.Diagnostics
	var data UrlModel

	if diags.Append(req.Plan.Get(ctx, &data)...); diags.HasError() {
		return
	}

	const errMessage = "Failed to update URL Origin"

	scheme, host, port, basePath := data.UrlModel.Parts(ctx)
	request := cdn77.OriginEditUrlJSONRequestBody{
		Label:   data.Label.ValueStringPointer(),
		Note:    util.StringValueToNullable(data.Note),
		Scheme:  util.Pointer(cdn77.OriginScheme(scheme)),
		Host:    &host,
		Port:    port,
		BaseDir: basePath,
	}

	response, err := r.Client.OriginEditUrlWithResponse(ctx, data.Id.ValueString(), request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ProcessEmptyResponse(diags, response, errMessage, func() {
		diags.Append(resp.State.Set(ctx, data)...)
	})
}

func (r *UrlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	diags := &resp.Diagnostics
	var data UrlModel

	if diags.Append(req.State.Get(ctx, &data)...); diags.HasError() {
		return
	}

	const errMessage = "Failed to delete URL Origin"

	response, err := r.Client.OriginDeleteUrlWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ValidateDeletionResponse(diags, response, errMessage)
}

func (*UrlResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *UrlResource) MoveState(context.Context) []resource.StateMover {
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

				if oldModel.Type.ValueString() != TypeUrl {
					diags.AddError(
						"Unable to Move Resource State",
						fmt.Sprintf(
							"Only %q resources with type=%q can be moved to %q.\nLabel: %s\nType: %s\n",
							DeprecatedOriginResourceType,
							TypeUrl,
							r.FullName(),
							oldModel.Label.ValueString(),
							oldModel.Type.ValueString(),
						),
					)

					return
				}

				model := UrlModel{
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
				}

				diags.Append(resp.TargetState.Set(ctx, model)...)
			},
		},
	}
}

var _ datasource.DataSourceWithConfigure = &UrlDataSource{}

type UrlDataSource struct {
	*util.BaseDataSource
}
