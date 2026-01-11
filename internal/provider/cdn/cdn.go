package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

	request, ok := r.createAddRequest(ctx, diags, data)
	if !ok {
		return
	}

	const errMessage = "Failed to create CDN"

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

	diags.Append(resp.State.Set(ctx, data)...)
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

type cdnRequestParams struct {
	Cache               *cdn77.Cache
	Cnames              *[]string
	ConditionalFeatures *cdn77.ConditionalFeatures
	GeoProtection       *cdn77.GeoProtection
	Headers             *cdn77.Headers
	HotlinkProtection   *cdn77.HotlinkProtection
	HttpsRedirect       *cdn77.HttpsRedirect
	IpProtection        *cdn77.IpProtection
	Mp4PseudoStreaming  *cdn77.Mp4PseudoStreaming
	Note                nullable.Nullable[string]
	OriginHeaders       *cdn77.OriginHeaders
	QueryString         *cdn77.QueryString
	RateLimit           *cdn77.RateLimit
	SecureToken         *cdn77.SecureToken
	Ssl                 *cdn77.CdnSsl
}

func (r *Resource) buildCdnRequestParams( //nolint:cyclop
	ctx context.Context,
	diags *diag.Diagnostics,
	data Model,
) (cdnRequestParams, bool) {
	params := cdnRequestParams{
		Cache: &cdn77.Cache{
			MaxAge:                     util.Pointer(cdn77.MaxAge(data.Cache.MaxAge.ValueInt64())),
			MaxAge404:                  util.Int64ValueToNullable[cdn77.MaxAge404](data.Cache.MaxAge404),
			RequestsWithCookiesEnabled: data.Cache.RequestsWithCookiesEnabled.ValueBoolPointer(),
		},
		GeoProtection: &cdn77.GeoProtection{
			Type: cdn77.AccessProtectionType(data.GeoProtection.Type.ValueString()),
		},
		Headers: &cdn77.Headers{
			ContentDisposition: &cdn77.ContentDisposition{
				Type: util.Pointer(cdn77.ContentDispositionType(data.Headers.ContentDispositionType.ValueString())),
			},
			CorsEnabled:                 data.Headers.CorsEnabled.ValueBoolPointer(),
			CorsTimingEnabled:           data.Headers.CorsTimingEnabled.ValueBoolPointer(),
			CorsWildcardEnabled:         data.Headers.CorsWildcardEnabled.ValueBoolPointer(),
			HostHeaderForwardingEnabled: data.Headers.HostHeaderForwardingEnabled.ValueBoolPointer(),
		},
		HotlinkProtection: &cdn77.HotlinkProtection{
			Type:               cdn77.AccessProtectionType(data.HotlinkProtection.Type.ValueString()),
			EmptyRefererDenied: data.HotlinkProtection.EmptyRefererDenied.ValueBool(),
		},
		HttpsRedirect: &cdn77.HttpsRedirect{
			Enabled: data.HttpsRedirect.Enabled.ValueBool(),
		},
		IpProtection: &cdn77.IpProtection{
			Type: cdn77.AccessProtectionType(data.IpProtection.Type.ValueString()),
		},
		Mp4PseudoStreaming: &cdn77.Mp4PseudoStreaming{
			Enabled: data.Mp4PseudoStreamingEnabled.ValueBoolPointer(),
		},
		Note: util.StringValueToNullable(data.Note),
		OriginHeaders: &cdn77.OriginHeaders{
			Custom: nullable.NewNullNullable[map[string]string](),
		},
		QueryString: &cdn77.QueryString{
			IgnoreType: cdn77.QueryStringIgnoreType(data.QueryString.IgnoreType.ValueString()),
		},
		RateLimit: &cdn77.RateLimit{
			Enabled: data.RateLimitEnabled.ValueBool(),
		},
		SecureToken: &cdn77.SecureToken{
			Type: cdn77.SecureTokenType(data.SecureToken.Type.ValueString()),
		},
		Ssl: &cdn77.CdnSsl{
			Type: cdn77.SslType(data.Ssl.Type.ValueString()),
		},
	}

	if !data.Cnames.IsNull() {
		cnames, ok := util.StringSetToSlice(ctx, diags, path.Root("cnames"), data.Cnames)
		if !ok {
			return cdnRequestParams{}, false
		}

		params.Cnames = &cnames
	}

	if !data.GeoProtection.Countries.IsNull() {
		countriesPath := path.Root("geo_protection").AtName("countries")
		countries, ok := util.StringSetToSlice(ctx, diags, countriesPath, data.GeoProtection.Countries)

		if !ok {
			return cdnRequestParams{}, false
		}

		params.GeoProtection.Countries = util.Pointer(countries)
	}

	if !data.HotlinkProtection.Domains.IsNull() {
		domainsPath := path.Root("hotlink_protection").AtName("domains")
		domains, ok := util.StringSetToSlice(ctx, diags, domainsPath, data.HotlinkProtection.Domains)

		if !ok {
			return cdnRequestParams{}, false
		}

		params.HotlinkProtection.Domains = util.Pointer(domains)
	}

	if !data.HttpsRedirect.Code.IsNull() {
		params.HttpsRedirect.Code = util.Pointer(cdn77.HttpsRedirectCode(data.HttpsRedirect.Code.ValueInt64()))
	}

	if !data.IpProtection.Ips.IsNull() {
		ips, ok := util.StringSetToSlice(ctx, diags, path.Root("ip_protection").AtName("ips"), data.IpProtection.Ips)
		if !ok {
			return cdnRequestParams{}, false
		}

		params.IpProtection.Ips = util.Pointer(ips)
	}

	if !data.OriginHeaders.IsNull() {
		headers, ok := util.StringMapToMap(ctx, diags, path.Root("origin_headers"), data.OriginHeaders)
		if !ok {
			return cdnRequestParams{}, false
		}

		params.OriginHeaders.Custom = nullable.NewNullableWithValue(headers)
	}

	if !data.QueryString.Parameters.IsNull() {
		queryParamsPath := path.Root("query_string").AtName("parameters")
		queryParams, ok := util.StringSetToSlice(ctx, diags, queryParamsPath, data.QueryString.Parameters)

		if !ok {
			return cdnRequestParams{}, false
		}

		params.QueryString.Parameters = util.Pointer(queryParams)
	}

	if !data.SecureToken.Token.IsNull() {
		params.SecureToken.Token = data.SecureToken.Token.ValueStringPointer()
	}

	if !data.Ssl.SslId.IsNull() {
		params.Ssl.SslId = data.Ssl.SslId.ValueStringPointer()
	}

	if data.ConditionalFeatures != nil {
		conditionalFeatures, ok := r.buildConditionalFeatures(diags, &data)
		if !ok {
			return cdnRequestParams{}, false
		}

		if conditionalFeatures != nil {
			params.ConditionalFeatures = conditionalFeatures
		}
	}

	return params, true
}

func (r *Resource) createAddRequest(
	ctx context.Context,
	diags *diag.Diagnostics,
	data Model,
) (cdn77.CdnAddJSONRequestBody, bool) {
	params, ok := r.buildCdnRequestParams(ctx, diags, data)
	if !ok {
		return cdn77.CdnAddJSONRequestBody{}, false
	}

	return cdn77.CdnAddJSONRequestBody{
		Label:               data.Label.ValueString(),
		OriginId:            data.OriginId.ValueString(),
		Cache:               params.Cache,
		Cnames:              params.Cnames,
		ConditionalFeatures: params.ConditionalFeatures,
		GeoProtection:       params.GeoProtection,
		Headers:             params.Headers,
		HotlinkProtection:   params.HotlinkProtection,
		HttpsRedirect:       params.HttpsRedirect,
		IpProtection:        params.IpProtection,
		Mp4PseudoStreaming:  params.Mp4PseudoStreaming,
		Note:                params.Note,
		OriginHeaders:       params.OriginHeaders,
		QueryString:         params.QueryString,
		RateLimit:           params.RateLimit,
		SecureToken:         params.SecureToken,
		Ssl:                 params.Ssl,
	}, true
}

func (r *Resource) createEditRequest(
	ctx context.Context,
	diags *diag.Diagnostics,
	data Model,
) (cdn77.CdnEditJSONRequestBody, bool) {
	params, ok := r.buildCdnRequestParams(ctx, diags, data)
	if !ok {
		return cdn77.CdnEditJSONRequestBody{}, false
	}

	return cdn77.CdnEditJSONRequestBody{
		Label:               data.Label.ValueStringPointer(),
		OriginId:            data.OriginId.ValueStringPointer(),
		Cache:               params.Cache,
		Cnames:              params.Cnames,
		ConditionalFeatures: params.ConditionalFeatures,
		GeoProtection:       params.GeoProtection,
		Headers:             params.Headers,
		HotlinkProtection:   params.HotlinkProtection,
		HttpsRedirect:       params.HttpsRedirect,
		IpProtection:        params.IpProtection,
		Mp4PseudoStreaming:  params.Mp4PseudoStreaming,
		Note:                params.Note,
		OriginHeaders:       params.OriginHeaders,
		QueryString:         params.QueryString,
		RateLimit:           params.RateLimit,
		SecureToken:         params.SecureToken,
		Ssl:                 params.Ssl,
	}, true
}

func (*Resource) buildConditionalFeatures(
	diags *diag.Diagnostics,
	data *Model,
) (*cdn77.ConditionalFeatures, bool) {
	if data.ConditionalFeatures == nil {
		return nil, true
	}

	conditionalFeatures := &cdn77.ConditionalFeatures{}

	hasConfig, ok := parseConditionalFeaturesConfig(diags, data, conditionalFeatures)
	if !ok {
		return nil, false
	}

	hasSecrets, ok := parseConditionalFeaturesSecrets(data, conditionalFeatures)
	if !ok {
		return nil, false
	}

	if !hasConfig && !hasSecrets {
		return nil, true
	}

	return conditionalFeatures, true
}

func parseConditionalFeaturesConfig(
	diags *diag.Diagnostics,
	data *Model,
	conditionalFeatures *cdn77.ConditionalFeatures,
) (hasConfig bool, ok bool) {
	configAttr := data.ConditionalFeatures.Configuration

	if configAttr.IsNull() || configAttr.IsUnknown() {
		return false, true
	}

	raw := strings.TrimSpace(configAttr.ValueString())
	if raw == "" {
		return false, true
	}

	normalizedBytes, err := normalizeConditionalFeaturesEmptyConfigArray([]byte(raw))
	if err != nil {
		diags.AddAttributeError(
			path.Root("conditional_features").AtName("configuration"),
			"Invalid conditional_features.configuration",
			fmt.Sprintf("Configuration must be a JSON array: %v", err),
		)

		return false, false
	}

	rawCanonical, err := canonicalizeJSON(string(normalizedBytes))
	if err != nil {
		diags.AddAttributeError(
			path.Root("conditional_features").AtName("configuration"),
			"Invalid JSON",
			fmt.Sprintf("Configuration must contain valid JSON: %v", err),
		)

		return false, false
	}

	var parsed []cdn77.ConditionalFeatureConfiguration
	if err := json.Unmarshal([]byte(rawCanonical), &parsed); err != nil {
		diags.AddAttributeError(
			path.Root("conditional_features").AtName("configuration"),
			"Invalid conditional_features.configuration",
			fmt.Sprintf("Configuration must be a JSON array: %v", err),
		)

		return false, false
	}

	if len(parsed) == 0 {
		return false, true
	}

	conditionalFeatures.Configuration = &parsed

	return true, true
}

func parseConditionalFeaturesSecrets(
	data *Model,
	conditionalFeatures *cdn77.ConditionalFeatures,
) (hasSecrets bool, ok bool) {
	secretAttr := data.ConditionalFeatures.Secrets

	if secretAttr.IsNull() || secretAttr.IsUnknown() {
		data.ConditionalFeatures.Secrets = types.MapNull(types.StringType)

		return false, true
	}

	elements := secretAttr.Elements()
	if len(elements) == 0 {
		return false, true
	}

	secrets := make(map[string]string)

	for key, valTyped := range elements {
		val := valTyped.(types.String)

		if val.IsNull() || val.IsUnknown() {
			continue
		}

		secrets[key] = val.ValueString()
	}

	if len(secrets) == 0 {
		return false, true
	}

	conditionalFeatures.Secrets = &secrets

	return true, true
}

func canonicalizeJSON(raw string) (string, error) {
	var data any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	canonicalJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(canonicalJSON), nil
}

var _ datasource.DataSourceWithConfigure = &DataSource{}

type DataSource struct {
	*util.BaseDataSource
}
