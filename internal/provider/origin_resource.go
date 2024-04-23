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
	case OriginTypeObjectStorage:
		if !r.createObjectStorage(ctx, &resp.Diagnostics, errMessage, &data) {
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

//nolint:cyclop
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
	case OriginTypeObjectStorage:
		var stateData OriginModel
		if resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...); resp.Diagnostics.HasError() {
			return
		}

		if !r.updateObjectStorage(ctx, &resp.Diagnostics, errMessage, &data, &stateData) {
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

	if data.Port.IsUnknown() {
		data.Port = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OriginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OriginModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	id := data.Id.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Can't delete Origin without ID", missingOriginIdDetailMessage)

		return
	}

	const errMessage = "Failed to delete Origin"

	switch data.Type.ValueString() {
	case OriginTypeAws:
		r.deleteAws(ctx, &resp.Diagnostics, &resp.State, errMessage, id)
	case OriginTypeObjectStorage:
		r.deleteObjectStorage(ctx, &resp.Diagnostics, &resp.State, errMessage, id)
	case OriginTypeUrl:
		r.deleteUrl(ctx, &resp.Diagnostics, &resp.State, errMessage, id)
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
	data.AccessKeyId = types.StringNull()
	data.AccessKeySecret = types.StringNull()
	data.Port = util.NullableIntToInt64Value(response.JSON201.Port)

	return true
}

func (r *OriginResource) createObjectStorage(
	ctx context.Context,
	diags *diag.Diagnostics,
	errMessage string,
	data *OriginModel,
) bool {
	request := cdn77.OriginAddObjectStorageJSONRequestBody{
		Acl:        cdn77.AclType(data.Acl.ValueString()),
		BucketName: data.BucketName.ValueString(),
		ClusterId:  data.ClusterId.ValueString(),
		Label:      data.Label.ValueString(),
		Note:       util.StringValueToNullable(data.Note),
	}

	response, err := r.client.OriginAddObjectStorageWithResponse(ctx, request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return false
	}

	if !util.CheckResponse(diags, errMessage, response, response.JSON422, response.JSONDefault) {
		return false
	}

	data.Id = types.StringValue(response.JSON201.Id)
	data.AccessKeyId = types.StringPointerValue(response.JSON201.AccessKeyId)
	data.AccessKeySecret = types.StringPointerValue(response.JSON201.AccessSecret)
	data.Scheme = types.StringValue(string(response.JSON201.Scheme))
	data.Host = types.StringValue(response.JSON201.Host)
	data.Port = util.NullableIntToInt64Value(response.JSON201.Port)

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
	data.AccessKeyId = types.StringNull()
	data.AccessKeySecret = types.StringNull()
	data.Port = util.NullableIntToInt64Value(response.JSON201.Port)

	return true
}

func (r *OriginResource) updateAws(
	ctx context.Context,
	diags *diag.Diagnostics,
	errMessage string,
	data *OriginModel,
) bool {
	data.AccessKeyId = types.StringNull()
	data.AccessKeySecret = types.StringNull()

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

func (r *OriginResource) updateObjectStorage(
	ctx context.Context,
	diags *diag.Diagnostics,
	errMessage string,
	data *OriginModel,
	stateData *OriginModel,
) bool {
	data.Scheme = stateData.Scheme
	data.Host = stateData.Host
	data.Port = stateData.Port

	request := cdn77.OriginEditObjectStorageJSONRequestBody{
		Label: data.Label.ValueStringPointer(),
		Note:  util.StringValueToNullable(data.Note),
	}

	response, err := r.client.OriginEditObjectStorageWithResponse(ctx, data.Id.ValueString(), request)
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
	data.AccessKeyId = types.StringNull()
	data.AccessKeySecret = types.StringNull()

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

func (r *OriginResource) deleteAws(
	ctx context.Context,
	diags *diag.Diagnostics,
	state *tfsdk.State,
	errMessage string,
	id string,
) {
	response, err := r.client.OriginDeleteAwsWithResponse(ctx, id)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	if maybeRemoveMissingResource(ctx, response.StatusCode(), id, state) {
		return
	}

	util.CheckResponse(diags, errMessage, response, response.JSON404, response.JSON422, response.JSONDefault)
}

func (r *OriginResource) deleteObjectStorage(
	ctx context.Context,
	diags *diag.Diagnostics,
	state *tfsdk.State,
	errMessage string,
	id string,
) {
	response, err := r.client.OriginDeleteObjectStorageWithResponse(ctx, id)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	if maybeRemoveMissingResource(ctx, response.StatusCode(), id, state) {
		return
	}

	util.CheckResponse(diags, errMessage, response, response.JSON404, response.JSON422, response.JSONDefault)
}

func (r *OriginResource) deleteUrl(
	ctx context.Context,
	diags *diag.Diagnostics,
	state *tfsdk.State,
	errMessage string,
	id string,
) {
	response, err := r.client.OriginDeleteUrlWithResponse(ctx, id)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	if maybeRemoveMissingResource(ctx, response.StatusCode(), id, state) {
		return
	}

	util.CheckResponse(diags, errMessage, response, response.JSON404, response.JSON422, response.JSONDefault)
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
