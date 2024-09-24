package origin

import (
	"context"
	"fmt"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type UrlReader struct{}

func (*UrlReader) ErrMessage() string {
	return "Failed to fetch URL Origin"
}

func (*UrlReader) Fetch(
	ctx context.Context,
	client cdn77.ClientWithResponsesInterface,
	model UrlModel,
) (*cdn77.OriginDetailUrlResponse, *cdn77.UrlOriginDetail, error) {
	response, err := client.OriginDetailUrlWithResponse(ctx, model.Id.ValueString())
	if err != nil {
		return nil, nil, err
	}

	return response, response.JSON200, nil
}

func (r *UrlReader) Process(
	ctx context.Context,
	model UrlModel,
	detail *cdn77.UrlOriginDetail,
	diags *diag.Diagnostics,
) UrlModel {
	if detail.Type != TypeUrl {
		diags.AddError(r.ErrMessage(), fmt.Sprintf("Origin with id=\"%s\" is not an URL Origin", detail.Id))

		return model
	}

	return UrlModel{
		SharedModel: NewSharedModel(model.Id, detail.Label, detail.Note),
		UrlModel:    shared.NewUrlModel(ctx, string(detail.Scheme), detail.Host, detail.Port, detail.BaseDir),
	}
}
