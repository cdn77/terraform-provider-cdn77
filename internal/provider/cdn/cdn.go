package cdn

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oapi-codegen/nullable"
)

var (
	_ resource.ResourceWithConfigure        = &Resource{}
	_ resource.ResourceWithConfigValidators = &Resource{}
	_ resource.ResourceWithImportState      = &Resource{}
)

type Resource struct {
	*util.BaseResource
}

func (*Resource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{NewNullableListsConfigValidator()}
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	diags := &resp.Diagnostics
	var data Model

	if diags.Append(req.Plan.Get(ctx, &data)...); diags.HasError() {
		return
	}

	var cnamesPtr *[]string

	if !data.Cnames.IsNull() {
		cnames, ok := util.StringSetToSlice(ctx, diags, path.Root("cnames"), data.Cnames)
		if !ok {
			return
		}

		cnamesPtr = &cnames
	}

	const errMessage = "Failed to create CDN"

	request := cdn77.CdnAddJSONRequestBody{
		OriginId: data.OriginId.ValueString(),
		Label:    data.Label.ValueString(),
		Cnames:   cnamesPtr,
		Note:     util.StringValueToNullable(data.Note),
	}

	response, err := r.Client.CdnAddWithResponse(ctx, request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	var id int

	util.ProcessResponse(diags, response, errMessage, response.JSON201, func(detail *cdn77.CdnSummary) {
		id = detail.Id
		data.Id = types.Int64Value(int64(id))
		data.CreationTime = types.StringValue(detail.CreationTime.Format(time.DateTime))
		data.Url = types.StringValue(detail.Url)
	})

	if diags.HasError() {
		return
	}

	editRequest, ok := r.createEditRequest(ctx, diags, data)
	if !ok {
		r.deleteAfterFailedEdit(ctx, diags, id)

		return
	}

	const editErrMessage = "Failed to edit CDN after creation"

	editResponse, err := r.Client.CdnEditWithResponse(ctx, response.JSON201.Id, editRequest)
	if err != nil {
		diags.AddError(editErrMessage, err.Error())
		r.deleteAfterFailedEdit(ctx, diags, id)

		return
	}

	util.ProcessEmptyResponse(diags, editResponse, errMessage, func() {
		diags.Append(resp.State.Set(ctx, data)...)
	})

	if diags.HasError() {
		r.deleteAfterFailedEdit(ctx, diags, id)
	}
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := &resp.Diagnostics
	var data Model

	if diags.Append(req.Plan.Get(ctx, &data)...); diags.HasError() {
		return
	}

	request, ok := r.createEditRequest(ctx, diags, data)
	if !ok {
		return
	}

	const errMessage = "Failed to update CDN"

	response, err := r.Client.CdnEditWithResponse(ctx, int(data.Id.ValueInt64()), request)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ProcessEmptyResponse(diags, response, errMessage, func() {
		diags.Append(resp.State.Set(ctx, data)...)
	})
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	diags := &resp.Diagnostics
	var data Model

	if diags.Append(req.State.Get(ctx, &data)...); diags.HasError() {
		return
	}

	const errMessage = "Failed to delete CDN"

	response, err := r.Client.CdnDeleteWithResponse(ctx, int(data.Id.ValueInt64()))
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
	id, err := strconv.ParseUint(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import Identifier",
			fmt.Sprintf("Expected unsigned integer, got: %q", req.ID),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *Resource) createEditRequest( //nolint:cyclop
	ctx context.Context,
	diags *diag.Diagnostics,
	data Model,
) (cdn77.CdnEditJSONRequestBody, bool) {
	request := r.createDefaultEditRequest()
	request.Label = data.Label.ValueStringPointer()
	request.OriginId = data.OriginId.ValueStringPointer()

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
	request.Mp4PseudoStreaming.Enabled = data.Mp4PseudoStreamingEnabled.ValueBoolPointer()
	request.Note = util.StringValueToNullable(data.Note)

	if !data.OriginHeaders.IsNull() {
		headers, ok := util.StringMapToMap(ctx, diags, path.Root("origin_headers"), data.OriginHeaders)
		if !ok {
			return cdn77.CdnEditJSONRequestBody{}, false
		}

		request.OriginHeaders.Custom = nullable.NewNullableWithValue(headers)
	}

	if !data.QueryString.Parameters.IsNull() {
		paramsPath := path.Root("query_string").AtName("parameters")
		params, ok := util.StringSetToSlice(ctx, diags, paramsPath, data.QueryString.Parameters)

		if !ok {
			return cdn77.CdnEditJSONRequestBody{}, false
		}

		request.QueryString.Parameters = util.Pointer(params)
	}

	request.QueryString.IgnoreType = cdn77.QueryStringIgnoreType(data.QueryString.IgnoreType.ValueString())

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

func (*Resource) createDefaultEditRequest() cdn77.CdnEditJSONRequestBody {
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
		RateLimit:          &cdn77.RateLimit{},
		SecureToken:        &cdn77.SecureToken{Type: cdn77.SecureTokenTypeNone},
		Ssl:                &cdn77.CdnSsl{Type: cdn77.InstantSsl},
		Waf:                &cdn77.Waf{},
	}
}

func (r *Resource) deleteAfterFailedEdit(ctx context.Context, diags *diag.Diagnostics, id int) {
	const errMessage = "Failed to remove CDN after failed edit"

	response, err := r.Client.CdnDeleteWithResponse(ctx, id)
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	util.ValidateDeletionResponse(diags, response, errMessage)
}

var _ datasource.DataSourceWithConfigure = &DataSource{}

type DataSource struct {
	*util.BaseDataSource
}
