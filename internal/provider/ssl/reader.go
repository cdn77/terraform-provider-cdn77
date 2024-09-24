package ssl

import (
	"context"
	"time"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Reader struct{}

func (*Reader) ErrMessage() string {
	return "Failed to fetch SSL"
}

func (*Reader) Fetch(
	ctx context.Context,
	client cdn77.ClientWithResponsesInterface,
	model Model,
) (*cdn77.SslSniDetailResponse, *cdn77.Ssl, error) {
	response, err := client.SslSniDetailWithResponse(ctx, model.Id.ValueString())
	if err != nil {
		return nil, nil, err
	}

	return response, response.JSON200, nil
}

func (*Reader) Process(ctx context.Context, model Model, ssl *cdn77.Ssl, diags *diag.Diagnostics) Model {
	readSslDetails(ctx, diags, &model.BaseModel, ssl)

	return model
}

func readSslDetails(ctx context.Context, diags *diag.Diagnostics, model *BaseModel, ssl *cdn77.Ssl) {
	model.Certificate = types.StringValue(ssl.Certificate)
	model.Subjects = util.SetValueFrom(ctx, diags, types.StringType, ssl.Cnames)
	model.ExpiresAt = types.StringValue(ssl.ExpiresAt.Format(time.DateTime))
}
