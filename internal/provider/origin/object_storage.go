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
	"github.com/oapi-codegen/nullable"
)

var (
	_ resource.ResourceWithConfigure   = &ObjectStorageResource{}
	_ resource.ResourceWithImportState = &ObjectStorageResource{}
	_ resource.ResourceWithMoveState   = &ObjectStorageResource{}
)

type ObjectStorageResource struct {
	*util.BaseResource
}

func (r *ObjectStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	diags := &resp.Diagnostics
	var data ObjectStorageModel

	if diags.Append(req.Plan.Get(ctx, &data)...); diags.HasError() {
		return
	}

	const errMessage = "Failed to create Object Storage Origin"

	request := cdn77.OriginCreateObjectStorageJSONRequestBody{
		Label:      data.Label.ValueString(),
		Note:       util.StringValueToNullable(data.Note),
		BucketName: data.BucketName.ValueString(),
		Acl:        cdn77.AclType(data.Acl.ValueString()),
		ClusterId:  data.ClusterId.ValueString(),
	}

	response, err := r.Client.OriginCreateObjectStorageWithResponse(ctx, request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ProcessResponse(diags, response, errMessage, response.JSON201, func(detail *cdn77.ObjectStorageOriginDetail) {
		data.Id = types.StringValue(detail.Id)
		data.UrlModel = shared.NewUrlModel(
			ctx,
			string(detail.Scheme),
			detail.Host,
			detail.Port,
			nullable.NewNullNullable[string](),
		)
		data.AccessKeyId = types.StringPointerValue(detail.AccessKeyId)
		data.AccessKeySecret = types.StringPointerValue(detail.AccessSecret)
		data.Usage = &ObjectStorageUsageModel{
			Files:     util.IntPointerToInt64Value(detail.Usage.FileCount),
			SizeBytes: util.IntPointerToInt64Value(detail.Usage.SizeBytes),
		}

		diags.Append(resp.State.Set(ctx, data)...)
	})
}

func (r *ObjectStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := &resp.Diagnostics
	var data ObjectStorageModel

	if diags.Append(req.Plan.Get(ctx, &data)...); diags.HasError() {
		return
	}

	const errMessage = "Failed to update Object Storage Origin"

	request := cdn77.OriginEditObjectStorageJSONRequestBody{
		Label: data.Label.ValueStringPointer(),
		Note:  util.StringValueToNullable(data.Note),
	}

	response, err := r.Client.OriginEditObjectStorageWithResponse(ctx, data.Id.ValueString(), request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ProcessEmptyResponse(diags, response, errMessage, func() {
		diags.Append(resp.State.Set(ctx, data)...)
	})
}

func (r *ObjectStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	diags := &resp.Diagnostics
	var data ObjectStorageModel

	if diags.Append(req.State.Get(ctx, &data)...); diags.HasError() {
		return
	}

	const errMessage = "Failed to delete Object Storage Origin"

	response, err := r.Client.OriginDeleteObjectStorageWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ValidateDeletionResponse(diags, response, errMessage)
}

func (*ObjectStorageResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 5 {
		resp.Diagnostics.AddError(
			"Invalid Import Identifier",
			fmt.Sprintf("Expected: <id>,<acl>,<cluster_id>,<access_key_id>,<access_key_secret>\nGot: %q", req.ID),
		)

		return
	}

	id, acl, clusterId, accessKeyId, accessKeySecret := idParts[0], idParts[1], idParts[2], idParts[3], idParts[4]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("acl"), acl)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), clusterId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("access_key_id"), accessKeyId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("access_key_secret"), accessKeySecret)...)
}

func (r *ObjectStorageResource) MoveState(context.Context) []resource.StateMover {
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

				if oldModel.Type.ValueString() != TypeObjectStorage {
					diags.AddError(
						"Unable to Move Resource State",
						fmt.Sprintf(
							"Only %q resources with type=%q can be moved to %q.\nLabel: %s\nType: %s\n",
							DeprecatedOriginResourceType,
							TypeObjectStorage,
							r.FullName(),
							oldModel.Label.ValueString(),
							oldModel.Type.ValueString(),
						),
					)

					return
				}

				model := ObjectStorageModel{
					ObjectStorageBaseModel: ObjectStorageBaseModel{
						SharedModel: NewSharedModel(
							oldModel.Id,
							oldModel.Label.ValueString(),
							util.StringValueToNullable(oldModel.Note),
						),
						BucketName: oldModel.BucketName,
					},
					Acl:             oldModel.Acl,
					ClusterId:       oldModel.ClusterId,
					AccessKeyId:     oldModel.AccessKeyId,
					AccessKeySecret: oldModel.AccessKeySecret,
				}

				diags.Append(resp.TargetState.Set(ctx, model)...)
			},
		},
	}
}

var _ datasource.DataSourceWithConfigure = &ObjectStorageDataSource{}

type ObjectStorageDataSource struct {
	*util.BaseDataSource
}
