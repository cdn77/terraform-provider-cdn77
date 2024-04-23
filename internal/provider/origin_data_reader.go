package provider

import (
	"context"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const missingOriginIdDetailMessage = "Origin ID is null, unknown or an empty string"

type OriginModel struct {
	Id                 types.String `tfsdk:"id"`
	Type               types.String `tfsdk:"type"`
	Label              types.String `tfsdk:"label"`
	Note               types.String `tfsdk:"note"`
	AwsAccessKeyId     types.String `tfsdk:"aws_access_key_id"`
	AwsAccessKeySecret types.String `tfsdk:"aws_access_key_secret"`
	AwsRegion          types.String `tfsdk:"aws_region"`
	Acl                types.String `tfsdk:"acl"`
	ClusterId          types.String `tfsdk:"cluster_id"`
	AccessKeyId        types.String `tfsdk:"access_key_id"`
	AccessKeySecret    types.String `tfsdk:"access_key_secret"`
	BucketName         types.String `tfsdk:"bucket_name"`
	Scheme             types.String `tfsdk:"scheme"`
	Host               types.String `tfsdk:"host"`
	Port               types.Int64  `tfsdk:"port"`
	BaseDir            types.String `tfsdk:"base_dir"`
}

type OriginDataReader struct {
	ctx                   context.Context
	client                cdn77.ClientWithResponsesInterface
	removeMissingResource bool
}

func NewOriginDataSourceReader(ctx context.Context, client cdn77.ClientWithResponsesInterface) *OriginDataReader {
	return &OriginDataReader{ctx: ctx, client: client, removeMissingResource: false}
}

func NewOriginResourceReader(ctx context.Context, client cdn77.ClientWithResponsesInterface) *OriginDataReader {
	return &OriginDataReader{ctx: ctx, client: client, removeMissingResource: true}
}

func (d *OriginDataReader) Read(provider StateProvider, diags *diag.Diagnostics, state *tfsdk.State) {
	var data OriginModel
	if diags.Append(provider.Get(d.ctx, &data)...); diags.HasError() {
		return
	}

	if data.Id.ValueString() == "" {
		diags.AddError("Can't fetch Origin without ID", missingOriginIdDetailMessage)

		return
	}

	var (
		ok         bool
		statusCode int
	)

	const errMessage = "Failed to fetch Origin"

	switch data.Type.ValueString() {
	case OriginTypeAws:
		ok, statusCode = d.readAws(diags, errMessage, &data)
	case OriginTypeObjectStorage:
		ok, statusCode = d.readObjectStorage(diags, errMessage, &data)
	case OriginTypeUrl:
		ok, statusCode = d.readUrl(diags, errMessage, &data)
	default:
		addUnknownOriginTypeError(diags, data)

		return
	}

	if ok {
		diags.Append(state.Set(d.ctx, &data)...)

		return
	}

	if d.removeMissingResource && maybeRemoveMissingResource(d.ctx, statusCode, data.Id.ValueString(), state) {
		return
	}
}

func (d *OriginDataReader) readAws(diags *diag.Diagnostics, message string, data *OriginModel) (bool, int) {
	response, err := d.client.OriginDetailAwsWithResponse(d.ctx, data.Id.ValueString())
	if err != nil {
		diags.AddError(message, err.Error())

		return false, 0
	}

	if !util.CheckResponse(diags, message, response, response.JSON404, response.JSONDefault) {
		return false, response.StatusCode()
	}

	*data = OriginModel{
		Id:                 data.Id,
		Type:               types.StringValue(OriginTypeAws),
		Label:              types.StringValue(response.JSON200.Label),
		Note:               util.NullableToStringValue(response.JSON200.Note),
		AwsAccessKeyId:     util.NullableToStringValue(response.JSON200.AwsAccessKeyId),
		AwsAccessKeySecret: data.AwsAccessKeySecret,
		AwsRegion:          util.NullableToStringValue(response.JSON200.AwsRegion),
		Scheme:             types.StringValue(string(response.JSON200.Scheme)),
		Host:               types.StringValue(response.JSON200.Host),
		Port:               util.NullableIntToInt64Value(response.JSON200.Port),
		BaseDir:            util.NullableToStringValue(response.JSON200.BaseDir),
	}

	return true, response.StatusCode()
}

func (d *OriginDataReader) readObjectStorage(diags *diag.Diagnostics, message string, data *OriginModel) (bool, int) {
	response, err := d.client.OriginDetailObjectStorageWithResponse(d.ctx, data.Id.ValueString())
	if err != nil {
		diags.AddError(message, err.Error())

		return false, 0
	}

	if !util.CheckResponse(diags, message, response, response.JSON404, response.JSONDefault) {
		return false, response.StatusCode()
	}

	*data = OriginModel{
		Id:              data.Id,
		Type:            types.StringValue(OriginTypeObjectStorage),
		Label:           types.StringValue(response.JSON200.Label),
		Note:            util.NullableToStringValue(response.JSON200.Note),
		Acl:             data.Acl,
		ClusterId:       data.ClusterId,
		AccessKeyId:     data.AccessKeyId,
		AccessKeySecret: data.AccessKeySecret,
		BucketName:      types.StringValue(response.JSON200.BucketName),
		Scheme:          types.StringValue(string(response.JSON200.Scheme)),
		Host:            types.StringValue(response.JSON200.Host),
		Port:            util.NullableIntToInt64Value(response.JSON200.Port),
	}

	return true, response.StatusCode()
}

func (d *OriginDataReader) readUrl(diags *diag.Diagnostics, message string, data *OriginModel) (bool, int) {
	response, err := d.client.OriginDetailUrlWithResponse(d.ctx, data.Id.ValueString())
	if err != nil {
		diags.AddError(message, err.Error())

		return false, 0
	}

	if !util.CheckResponse(diags, message, response, response.JSON404, response.JSONDefault) {
		return false, response.StatusCode()
	}

	*data = OriginModel{
		Id:      data.Id,
		Type:    types.StringValue(OriginTypeUrl),
		Label:   types.StringValue(response.JSON200.Label),
		Note:    util.NullableToStringValue(response.JSON200.Note),
		Scheme:  types.StringValue(string(response.JSON200.Scheme)),
		Host:    types.StringValue(response.JSON200.Host),
		Port:    util.NullableIntToInt64Value(response.JSON200.Port),
		BaseDir: util.NullableToStringValue(response.JSON200.BaseDir),
	}

	return true, response.StatusCode()
}
