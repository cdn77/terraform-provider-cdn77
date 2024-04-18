package provider

import (
	"context"
	"time"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const missingSslIdDetailMessage = "SSL ID is null, unknown or an empty string"

type SslModel struct {
	Id          types.String `tfsdk:"id"`
	Certificate types.String `tfsdk:"certificate"`
	PrivateKey  types.String `tfsdk:"private_key"`
	Subjects    types.Set    `tfsdk:"subjects"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
}

type SslDataReader struct {
	ctx                   context.Context
	client                cdn77.ClientWithResponsesInterface
	removeMissingResource bool
}

func NewSslDataSourceReader(ctx context.Context, client cdn77.ClientWithResponsesInterface) *SslDataReader {
	return &SslDataReader{ctx: ctx, client: client, removeMissingResource: false}
}

func NewSslResourceReader(ctx context.Context, client cdn77.ClientWithResponsesInterface) *SslDataReader {
	return &SslDataReader{ctx: ctx, client: client, removeMissingResource: true}
}

func (d *SslDataReader) Read(provider StateProvider, diags *diag.Diagnostics, state *tfsdk.State) {
	var data SslModel
	if diags.Append(provider.Get(d.ctx, &data)...); diags.HasError() {
		return
	}

	id := data.Id.ValueString()
	if id == "" {
		diags.AddError("Can't fetch SSL without ID", missingSslIdDetailMessage)

		return
	}

	const errMessage = "Failed to fetch SSL"

	response, err := d.client.SslSniDetailWithResponse(d.ctx, id)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	if d.removeMissingResource && maybeRemoveMissingResource(d.ctx, response.StatusCode(), id, state) {
		return
	}

	if !util.CheckResponse(diags, errMessage, response, response.JSON404, response.JSONDefault) {
		return
	}

	if ds := readSslDetails(d.ctx, &data, response.JSON200); ds != nil {
		diags.Append(ds...)

		return
	}

	diags.Append(state.Set(d.ctx, &data)...)
}

func readSslDetails(ctx context.Context, data *SslModel, ssl *cdn77.Ssl) diag.Diagnostics {
	data.Certificate = types.StringValue(ssl.Certificate)

	subjects, ds := types.SetValueFrom(ctx, types.StringType, ssl.Cnames)
	if ds != nil {
		return ds
	}

	data.Subjects = subjects
	data.ExpiresAt = types.StringValue(ssl.ExpiresAt.Format(time.DateTime))

	return nil
}
