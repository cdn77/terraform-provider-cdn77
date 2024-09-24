package ssl

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type Resource struct {
	*util.BaseResource
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	diags := &resp.Diagnostics
	var data Model

	if diags.Append(req.Plan.Get(ctx, &data)...); diags.HasError() {
		return
	}

	request := cdn77.SslSniAddJSONRequestBody{
		Certificate: data.Certificate.ValueString(),
		PrivateKey:  data.PrivateKey.ValueString(),
	}

	const errMessage = "Failed to create SSL"

	response, err := r.Client.SslSniAddWithResponse(ctx, request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ProcessResponse(diags, response, errMessage, response.JSON201, func(ssl *cdn77.Ssl) {
		data.Id = types.StringValue(ssl.Id)
		if readSslDetails(ctx, diags, &data.BaseModel, ssl); diags.HasError() {
			return
		}

		diags.Append(resp.State.Set(ctx, data)...)
	})
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := &resp.Diagnostics
	var data Model

	if diags.Append(req.Plan.Get(ctx, &data)...); diags.HasError() {
		return
	}

	request := cdn77.SslSniEditJSONRequestBody{
		Certificate: data.Certificate.ValueString(),
		PrivateKey:  data.PrivateKey.ValueStringPointer(),
	}

	const errMessage = "Failed to update SSL"

	response, err := r.Client.SslSniEditWithResponse(ctx, data.Id.ValueString(), request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ProcessResponse(diags, response, errMessage, response.JSON200, func(ssl *cdn77.Ssl) {
		if readSslDetails(ctx, diags, &data.BaseModel, ssl); diags.HasError() {
			return
		}

		diags.Append(resp.State.Set(ctx, data)...)
	})
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	diags := &resp.Diagnostics
	var data Model

	if diags.Append(req.State.Get(ctx, &data)...); diags.HasError() {
		return
	}

	const errMessage = "Failed to delete SSL"

	response, err := r.Client.SslSniDeleteWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ValidateDeletionResponse(diags, response, errMessage)
}

func (*Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf(
				"Expected import identifier with format: <id>,<privateKey>. "+
					"<privateKey> must be the whole PEM file (including headers) encoded via base64. Got: %q",
				req.ID,
			),
		)

		return
	}

	id, privateKeyBase64 := idParts[0], idParts[1]

	privateKey, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Private Key",
			"Private Key must be base64 encoded key (including the PEM headers)")

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("private_key"), string(privateKey))...)
}

var _ datasource.DataSourceWithConfigure = &DataSource{}

type DataSource struct {
	*util.BaseDataSource
}
