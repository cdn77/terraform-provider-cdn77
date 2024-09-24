package origin

import (
	"context"
	"fmt"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/shared"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type AwsReader struct{}

func (*AwsReader) ErrMessage() string {
	return "Failed to fetch AWS Origin"
}

func (*AwsReader) Fetch(
	ctx context.Context,
	client cdn77.ClientWithResponsesInterface,
	model AwsModel,
) (*cdn77.OriginDetailAwsResponse, *cdn77.S3OriginDetail, error) {
	response, err := client.OriginDetailAwsWithResponse(ctx, model.Id.ValueString())
	if err != nil {
		return nil, nil, err
	}

	return response, response.JSON200, nil
}

func (r *AwsReader) Process(
	ctx context.Context,
	model AwsModel,
	detail *cdn77.S3OriginDetail,
	diags *diag.Diagnostics,
) AwsModel {
	if detail.Type != TypeAws {
		diags.AddError(r.ErrMessage(), fmt.Sprintf("Origin with id=\"%s\" is not an AWS Origin", detail.Id))

		return model
	}

	return AwsModel{
		AwsBaseModel: AwsBaseModel{
			SharedModel: NewSharedModel(model.Id, detail.Label, detail.Note),
			UrlModel:    shared.NewUrlModel(ctx, string(detail.Scheme), detail.Host, detail.Port, detail.BaseDir),
			AccessKeyId: util.NullableToStringValue(detail.AwsAccessKeyId),
			Region:      util.NullableToStringValue(detail.AwsRegion),
		},
		AccessKeySecret: model.AccessKeySecret,
	}
}
