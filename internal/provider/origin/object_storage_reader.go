package origin

import (
	"context"
	"fmt"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/shared"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oapi-codegen/nullable"
)

type ObjectStorageReader struct{}

func (*ObjectStorageReader) ErrMessage() string {
	return "Failed to fetch Object Storage Origin"
}

func (*ObjectStorageReader) Fetch(
	ctx context.Context,
	client cdn77.ClientWithResponsesInterface,
	model ObjectStorageModel,
) (*cdn77.OriginDetailObjectStorageResponse, *cdn77.ObjectStorageOriginDetail, error) {
	response, err := client.OriginDetailObjectStorageWithResponse(ctx, model.Id.ValueString())
	if err != nil {
		return response, nil, err
	}

	return response, response.JSON200, nil
}

func (r *ObjectStorageReader) Process(
	ctx context.Context,
	model ObjectStorageModel,
	detail *cdn77.ObjectStorageOriginDetail,
	diags *diag.Diagnostics,
) ObjectStorageModel {
	if detail.Type != TypeObjectStorage {
		diags.AddError(r.ErrMessage(), fmt.Sprintf("Origin with id=\"%s\" is not an Object Storage Origin", detail.Id))

		return model
	}

	return ObjectStorageModel{
		ObjectStorageBaseModel: ObjectStorageBaseModel{
			SharedModel: NewSharedModel(model.Id, detail.Label, detail.Note),
			UrlModel: shared.NewUrlModel(
				ctx,
				string(detail.Scheme),
				detail.Host,
				detail.Port,
				nullable.NewNullNullable[string]()),
			BucketName: types.StringValue(detail.BucketName),
			Usage: &ObjectStorageUsageModel{
				Files:     util.IntPointerToInt64Value(detail.Usage.FileCount),
				SizeBytes: util.IntPointerToInt64Value(detail.Usage.SizeBytes),
			},
		},
		Acl:             model.Acl,
		ClusterId:       model.ClusterId,
		AccessKeyId:     model.AccessKeyId,
		AccessKeySecret: model.AccessKeySecret,
	}
}
