package provider

import (
	"context"
	"time"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oapi-codegen/nullable"
)

var (
	_ resource.ResourceWithConfigure        = &CdnResource{}
	_ resource.ResourceWithConfigValidators = &CdnResource{}
)

func NewCdnResource() resource.Resource {
	return &CdnResource{}
}

type CdnResource struct {
	client cdn77.ClientWithResponsesInterface
}

func (*CdnResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cdn"
}

func (*CdnResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CreateCdnResourceSchema()
}

func (r *CdnResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(util.MaybeSetClient(req.ProviderData, &r.client))
}

func (*CdnResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{NewCdnNullableListsConfigValidator()}
}

func (r *CdnResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CdnModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	var cnamesPtr *[]string

	if !data.Cnames.IsNull() {
		cnames, ok := util.StringSetToSlice(ctx, &resp.Diagnostics, path.Root("cnames"), data.Cnames)
		if !ok {
			return
		}

		cnamesPtr = &cnames
	}

	const errMessage = "Failed to create CDN"

	request := cdn77.CdnAddJSONRequestBody{
		Cnames:   cnamesPtr,
		Label:    data.Label.ValueString(),
		Note:     util.StringValueToNullable(data.Note),
		OriginId: data.OriginId.ValueString(),
	}

	response, err := r.client.CdnAddWithResponse(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(errMessage, err.Error())

		return
	}

	if !util.CheckResponse(&resp.Diagnostics, errMessage, response, response.JSON422, response.JSONDefault) {
		return
	}

	data.Id = types.Int64Value(int64(response.JSON201.Id))
	data.CreationTime = types.StringValue(response.JSON201.CreationTime.Format(time.DateTime))
	data.OriginProtectionEnabled = types.BoolValue(false)
	data.Url = types.StringValue(response.JSON201.Url)

	editRequest, ok := r.createEditRequest(ctx, &resp.Diagnostics, data)
	if !ok {
		return
	}

	const editErrMessage = "Failed to edit CDN after creation; going to remove it"

	editResponse, err := r.client.CdnEditWithResponse(ctx, response.JSON201.Id, editRequest)
	if err == nil {
		if util.CheckResponse(
			&resp.Diagnostics,
			editErrMessage,
			editResponse,
			editResponse.JSON404,
			editResponse.JSON422,
			editResponse.JSONDefault,
		) {
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

			return
		}
	} else {
		resp.Diagnostics.AddError(editErrMessage, err.Error())
	}

	const deleteErrMessage = "Failed to remove CDN after failed edit"

	deleteResponse, err := r.client.CdnDeleteWithResponse(ctx, response.JSON201.Id)
	if err != nil {
		resp.Diagnostics.AddError(deleteErrMessage, err.Error())

		return
	}

	util.CheckResponse(
		&resp.Diagnostics,
		deleteErrMessage,
		deleteResponse,
		deleteResponse.JSON404,
		deleteResponse.JSONDefault,
	)
}

func (r *CdnResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	NewCdnResourceReader(ctx, r.client).Read(&req.State, &resp.Diagnostics, &resp.State)
}

func (r *CdnResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CdnModel
	if resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	request, ok := r.createEditRequest(ctx, &resp.Diagnostics, data)
	if !ok {
		return
	}

	const errMessage = "Failed to update CDN"

	response, err := r.client.CdnEditWithResponse(ctx, int(data.Id.ValueInt64()), request)
	if err != nil {
		resp.Diagnostics.AddError(errMessage, err.Error())

		return
	}

	if util.CheckResponse(
		&resp.Diagnostics,
		errMessage,
		response,
		response.JSON404,
		response.JSON422,
		response.JSONDefault,
	) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *CdnResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CdnModel
	if resp.Diagnostics.Append(req.State.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	const errMessage = "Failed to delete CDN"

	response, err := r.client.CdnDeleteWithResponse(ctx, int(data.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(errMessage, err.Error())

		return
	}

	if maybeRemoveMissingResource(ctx, response.StatusCode(), data.Id.ValueInt64(), &resp.State) {
		return
	}

	util.CheckResponse(&resp.Diagnostics, errMessage, response, response.JSON404, response.JSONDefault)
}

func (*CdnResource) createDefaultEditRequest() cdn77.CdnEditJSONRequestBody {
	return cdn77.CdnEditJSONRequestBody{
		Cache: &cdn77.Cache{
			MaxAge:                     util.Pointer(cdn77.MaxAgeN17280),
			MaxAge404:                  nullable.NewNullNullable[cdn77.MaxAge404](),
			RequestsWithCookiesEnabled: util.Pointer(true),
		},
		GeoProtection: &cdn77.GeoProtection{Type: cdn77.Disabled},
		Headers: &cdn77.Headers{
			ContentDisposition: &cdn77.ContentDisposition{Type: util.Pointer(cdn77.ContentDispositionTypeNone)},
		},
		HotlinkProtection:  &cdn77.HotlinkProtection{Type: cdn77.Disabled},
		HttpsRedirect:      &cdn77.HttpsRedirect{},
		IpProtection:       &cdn77.IpProtection{Type: cdn77.Disabled},
		Mp4PseudoStreaming: &cdn77.Mp4PseudoStreaming{Enabled: util.Pointer(false)},
		Note:               nullable.NewNullNullable[string](),
		OriginHeaders:      &cdn77.OriginHeaders{Custom: nullable.NewNullNullable[map[string]string]()},
		QueryString:        &cdn77.QueryString{IgnoreType: cdn77.QueryStringIgnoreTypeNone},
		Quic:               &cdn77.Quic{},
		RateLimit:          &cdn77.RateLimit{},
		SecureToken:        &cdn77.SecureToken{Type: cdn77.SecureTokenTypeNone},
		Ssl:                &cdn77.CdnSsl{Type: cdn77.InstantSsl},
		Waf:                &cdn77.Waf{},
	}
}

func (r *CdnResource) createEditRequest( //nolint:cyclop
	ctx context.Context,
	diags *diag.Diagnostics,
	data CdnModel,
) (cdn77.CdnEditJSONRequestBody, bool) {
	request := r.createDefaultEditRequest()

	request.Cache.MaxAge = util.Pointer(cdn77.MaxAge(data.Cache.MaxAge.ValueInt64()))
	request.Cache.MaxAge404 = util.Int64ValueToNullable[cdn77.MaxAge404](data.Cache.MaxAge404)
	request.Cache.RequestsWithCookiesEnabled = data.Cache.RequestsWithCookiesEnabled.ValueBoolPointer()

	if !data.Cnames.IsNull() {
		cnames, ok := util.StringSetToSlice(ctx, diags, path.Root("cnames"), data.Cnames)
		if !ok {
			return cdn77.CdnEditJSONRequestBody{}, false
		}

		request.Cnames = &cnames
	}

	if !data.GeoProtection.Countries.IsNull() {
		countriesPath := path.Root("geo_protection").AtName("countries")
		countries, ok := util.StringSetToSlice(ctx, diags, countriesPath, data.GeoProtection.Countries)

		if !ok {
			return cdn77.CdnEditJSONRequestBody{}, false
		}

		request.GeoProtection.Countries = util.Pointer(countries)
	}

	request.GeoProtection.Type = cdn77.AccessProtectionType(data.GeoProtection.Type.ValueString())

	request.Headers.ContentDisposition.Type = util.Pointer(
		cdn77.ContentDispositionType(data.Headers.ContentDispositionType.ValueString()),
	)
	request.Headers.CorsEnabled = data.Headers.CorsEnabled.ValueBoolPointer()
	request.Headers.CorsTimingEnabled = data.Headers.CorsTimingEnabled.ValueBoolPointer()
	request.Headers.CorsWildcardEnabled = data.Headers.CorsWildcardEnabled.ValueBoolPointer()
	request.Headers.HostHeaderForwardingEnabled = data.Headers.HostHeaderForwardingEnabled.ValueBoolPointer()

	if !data.HotlinkProtection.Domains.IsNull() {
		domainsPath := path.Root("hotlink_protection").AtName("domains")
		domains, ok := util.StringSetToSlice(ctx, diags, domainsPath, data.HotlinkProtection.Domains)

		if !ok {
			return cdn77.CdnEditJSONRequestBody{}, false
		}

		request.HotlinkProtection.Domains = util.Pointer(domains)
	}

	request.HotlinkProtection.Type = cdn77.AccessProtectionType(data.HotlinkProtection.Type.ValueString())
	request.HotlinkProtection.EmptyRefererDenied = data.HotlinkProtection.EmptyRefererDenied.ValueBool()

	if !data.HttpsRedirect.Code.IsNull() {
		request.HttpsRedirect.Code = util.Pointer(cdn77.HttpsRedirectCode(data.HttpsRedirect.Code.ValueInt64()))
	}

	request.HttpsRedirect.Enabled = data.HttpsRedirect.Enabled.ValueBool()

	if !data.IpProtection.Ips.IsNull() {
		ips, ok := util.StringSetToSlice(ctx, diags, path.Root("ip_protection").AtName("ips"), data.IpProtection.Ips)
		if !ok {
			return cdn77.CdnEditJSONRequestBody{}, false
		}

		request.IpProtection.Ips = util.Pointer(ips)
	}

	request.IpProtection.Type = cdn77.AccessProtectionType(data.IpProtection.Type.ValueString())

	request.Label = data.Label.ValueStringPointer()
	request.Mp4PseudoStreaming.Enabled = data.Mp4PseudoStreamingEnabled.ValueBoolPointer()
	request.Note = util.StringValueToNullable(data.Note)

	if !data.OriginHeaders.IsNull() {
		headers, ok := util.StringMapToMap(ctx, diags, path.Root("origin_headers"), data.OriginHeaders)
		if !ok {
			return cdn77.CdnEditJSONRequestBody{}, false
		}

		request.OriginHeaders.Custom = nullable.NewNullableWithValue(headers)
	}

	request.OriginId = data.OriginId.ValueStringPointer()

	if !data.QueryString.Parameters.IsNull() {
		paramsPath := path.Root("query_string").AtName("parameters")
		params, ok := util.StringSetToSlice(ctx, diags, paramsPath, data.QueryString.Parameters)

		if !ok {
			return cdn77.CdnEditJSONRequestBody{}, false
		}

		request.QueryString.Parameters = util.Pointer(params)
	}

	request.QueryString.IgnoreType = cdn77.QueryStringIgnoreType(data.QueryString.IgnoreType.ValueString())

	request.Quic.Enabled = data.QuicEnabled.ValueBool()
	request.RateLimit.Enabled = data.RateLimitEnabled.ValueBool()

	if !data.SecureToken.Token.IsNull() {
		request.SecureToken.Token = data.SecureToken.Token.ValueStringPointer()
	}

	request.SecureToken.Type = cdn77.SecureTokenType(data.SecureToken.Type.ValueString())

	if !data.Ssl.SslId.IsNull() {
		request.Ssl.SslId = data.Ssl.SslId.ValueStringPointer()
	}

	request.Ssl.Type = cdn77.SslType(data.Ssl.Type.ValueString())

	request.Waf.Enabled = data.WafEnabled.ValueBool()

	return request, true
}
