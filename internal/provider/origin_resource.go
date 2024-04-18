package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.ResourceWithConfigure        = &OriginResource{}
	_ resource.ResourceWithConfigValidators = &OriginResource{}
)

func NewOriginResource() resource.Resource {
	return &OriginResource{}
}

type OriginResource struct {
	client cdn77.ClientWithResponsesInterface
}

func (*OriginResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_origin"
}

func (*OriginResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CreateOriginResourceSchema()
}

func (r *OriginResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(util.MaybeSetClient(req.ProviderData, &r.client))
}

func (*OriginResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{NewOriginTypeConfigValidator()}
}

func (r *OriginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OriginModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	const errMessage = "Failed to create Origin"

	switch data.Type.ValueString() {
	case OriginTypeAws:
		if !r.createAws(ctx, &resp.Diagnostics, errMessage, &data) {
			return
		}
	case OriginTypeUrl:
		if !r.createUrl(ctx, &resp.Diagnostics, errMessage, &data) {
			return
		}
	default:
		addUnknownOriginTypeError(&resp.Diagnostics, data)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OriginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	NewOriginResourceReader(ctx, r.client).Read(&req.State, &resp.Diagnostics, &resp.State)
}

func (r *OriginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OriginModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	const errMessage = "Failed to update Origin"

	switch data.Type.ValueString() {
	case OriginTypeAws:
		if !r.updateAws(ctx, &resp.Diagnostics, errMessage, &data) {
			return
		}
	case OriginTypeUrl:
		if !r.updateUrl(ctx, &resp.Diagnostics, errMessage, &data) {
			return
		}
	default:
		addUnknownOriginTypeError(&resp.Diagnostics, data)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OriginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OriginModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	if data.Id.ValueString() == "" {
		resp.Diagnostics.AddError("Can't delete Origin without ID", missingOriginIdDetailMessage)

		return
	}

	const errMessage = "Failed to delete Origin"

	switch data.Type.ValueString() {
	case OriginTypeAws:
		response, err := r.client.OriginDeleteAwsWithResponse(ctx, data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(errMessage, err.Error())

			return
		}

		if maybeRemoveMissingResource(ctx, response.StatusCode(), data.Id.ValueString(), &resp.State) {
			return
		}

		util.CheckResponse(
			&resp.Diagnostics,
			errMessage,
			response,
			response.JSON404,
			response.JSON422,
			response.JSONDefault,
		)
	case OriginTypeUrl:
		response, err := r.client.OriginDeleteUrlWithResponse(ctx, data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(errMessage, err.Error())

			return
		}

		if maybeRemoveMissingResource(ctx, response.StatusCode(), data.Id.ValueString(), &resp.State) {
			return
		}

		util.CheckResponse(
			&resp.Diagnostics,
			errMessage,
			response,
			response.JSON404,
			response.JSON422,
			response.JSONDefault,
		)
	default:
		addUnknownOriginTypeError(&resp.Diagnostics, data)
	}
}

func (r *OriginResource) createAws(
	ctx context.Context,
	diags *diag.Diagnostics,
	errMessage string,
	data *OriginModel,
) bool {
	request := cdn77.OriginAddAwsJSONRequestBody{
		AwsAccessKeyId:     util.StringValueToNullable(data.AwsAccessKeyId),
		AwsAccessKeySecret: util.StringValueToNullable(data.AwsAccessKeySecret),
		AwsRegion:          util.StringValueToNullable(data.AwsRegion),
		BaseDir:            util.StringValueToNullable(data.BaseDir),
		Host:               data.Host.ValueString(),
		Label:              data.Label.ValueString(),
		Note:               util.StringValueToNullable(data.Note),
		Port:               util.Int64ValueToNullable[int](data.Port),
		Scheme:             cdn77.OriginScheme(data.Scheme.ValueString()),
	}

	response, err := r.client.OriginAddAwsWithResponse(ctx, request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return false
	}

	if !util.CheckResponse(diags, errMessage, response, response.JSON422, response.JSONDefault) {
		return false
	}

	data.Id = types.StringValue(response.JSON201.Id)

	return true
}

func (r *OriginResource) createUrl(
	ctx context.Context,
	diags *diag.Diagnostics,
	errMessage string,
	data *OriginModel,
) bool {
	request := cdn77.OriginAddUrlJSONRequestBody{
		BaseDir: util.StringValueToNullable(data.BaseDir),
		Host:    data.Host.ValueString(),
		Label:   data.Label.ValueString(),
		Note:    util.StringValueToNullable(data.Note),
		Port:    util.Int64ValueToNullable[int](data.Port),
		Scheme:  cdn77.OriginScheme(data.Scheme.ValueString()),
	}

	response, err := r.client.OriginAddUrlWithResponse(ctx, request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return false
	}

	if !util.CheckResponse(diags, errMessage, response, response.JSON422, response.JSONDefault) {
		return false
	}

	data.Id = types.StringValue(response.JSON201.Id)

	return true
}

func (r *OriginResource) updateAws(
	ctx context.Context,
	diags *diag.Diagnostics,
	errMessage string,
	data *OriginModel,
) bool {
	request := cdn77.OriginEditAwsJSONRequestBody{
		AwsAccessKeyId:     util.StringValueToNullable(data.AwsAccessKeyId),
		AwsAccessKeySecret: util.StringValueToNullable(data.AwsAccessKeySecret),
		AwsRegion:          util.StringValueToNullable(data.AwsRegion),
		BaseDir:            util.StringValueToNullable(data.BaseDir),
		Host:               data.Host.ValueStringPointer(),
		Label:              data.Label.ValueStringPointer(),
		Note:               util.StringValueToNullable(data.Note),
		Port:               util.Int64ValueToNullable[int](data.Port),
		Scheme:             data.Scheme.ValueStringPointer(),
	}

	response, err := r.client.OriginEditAwsWithResponse(ctx, data.Id.ValueString(), request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return false
	}

	return util.CheckResponse(diags, errMessage, response, response.JSON404, response.JSON422, response.JSONDefault)
}

func (r *OriginResource) updateUrl(
	ctx context.Context,
	diags *diag.Diagnostics,
	errMessage string,
	data *OriginModel,
) bool {
	request := cdn77.OriginEditUrlJSONRequestBody{
		BaseDir: util.StringValueToNullable(data.BaseDir),
		Host:    data.Host.ValueStringPointer(),
		Label:   data.Label.ValueStringPointer(),
		Note:    util.StringValueToNullable(data.Note),
		Port:    util.Int64ValueToNullable[int](data.Port),
		Scheme:  data.Scheme.ValueStringPointer(),
	}

	response, err := r.client.OriginEditUrlWithResponse(ctx, data.Id.ValueString(), request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return false
	}

	return util.CheckResponse(diags, errMessage, response, response.JSON404, response.JSON422, response.JSONDefault)
}

func maybeRemoveMissingResource(ctx context.Context, statusCode int, id any, state *tfsdk.State) bool {
	if statusCode == http.StatusNotFound {
		tflog.Info(tflog.SetField(ctx, "id", id), "Resource not found; removing it from state")
		state.RemoveResource(ctx)

		return true
	}

	return false
}

func addUnknownOriginTypeError(diags *diag.Diagnostics, data OriginModel) {
	diags.AddError(
		"Unknown Origin type",
		fmt.Sprintf(`Got "%s", expected one of: %v`, data.Type.ValueString(), originTypes),
	)
}
