package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &SslResource{}
	_ resource.ResourceWithImportState = &SslResource{}
)

func NewSslResource() resource.Resource {
	return &SslResource{}
}

type SslResource struct {
	client cdn77.ClientWithResponsesInterface
}

func (*SslResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssl"
}

func (*SslResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CreateSslResourceSchema()
}

func (r *SslResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(util.MaybeSetClient(req.ProviderData, &r.client))
}

func (r *SslResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SslModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	request := cdn77.SslSniAddJSONRequestBody{
		Certificate: data.Certificate.ValueString(),
		PrivateKey:  data.PrivateKey.ValueString(),
	}

	const errMessage = "Failed to create SSL"

	response, err := r.client.SslSniAddWithResponse(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(errMessage, err.Error())

		return
	}

	if !util.CheckResponse(&resp.Diagnostics, errMessage, response, response.JSON422, response.JSONDefault) {
		return
	}

	data.Id = types.StringValue(response.JSON201.Id)

	if ds := readSslDetails(ctx, &data, response.JSON201); ds != nil {
		resp.Diagnostics.Append(ds...)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SslResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	NewSslResourceReader(ctx, r.client).Read(&req.State, &resp.Diagnostics, &resp.State)
}

func (r *SslResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SslModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	request := cdn77.SslSniEditJSONRequestBody{
		Certificate: data.Certificate.ValueString(),
		PrivateKey:  data.PrivateKey.ValueStringPointer(),
	}

	const errMessage = "Failed to update SSL"

	response, err := r.client.SslSniEditWithResponse(ctx, data.Id.ValueString(), request)
	if err != nil {
		resp.Diagnostics.AddError(errMessage, err.Error())

		return
	}

	if !util.CheckResponse(&resp.Diagnostics, errMessage, response, response.JSON422, response.JSONDefault) {
		return
	}

	if ds := readSslDetails(ctx, &data, response.JSON200); ds != nil {
		resp.Diagnostics.Append(ds...)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SslResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SslModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	if data.Id.ValueString() == "" {
		resp.Diagnostics.AddError("Can't delete SSL without ID", missingSslIdDetailMessage)

		return
	}

	const errMessage = "Failed to delete SSL"

	response, err := r.client.SslSniDeleteWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errMessage, err.Error())

		return
	}

	if maybeRemoveMissingResource(ctx, response.StatusCode(), data.Id.ValueString(), &resp.State) {
		return
	}

	if !util.CheckResponse(&resp.Diagnostics, errMessage, response, response.JSON404, response.JSONDefault) {
		return
	}
}

func (*SslResource) ImportState(
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
