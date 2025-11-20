package cdn

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Reader struct{}

func (*Reader) ErrMessage() string {
	return "Failed to fetch CDN"
}

func (*Reader) Fetch(
	ctx context.Context,
	client cdn77.ClientWithResponsesInterface,
	model Model,
) (*cdn77.CdnDetailResponse, *cdn77.Cdn, error) {
	response, err := client.CdnDetailWithResponse(ctx, int(model.Id.ValueInt64()))
	if err != nil {
		return nil, nil, err
	}

	return response, response.JSON200, nil
}

func (r *Reader) Process(ctx context.Context, model Model, cdn *cdn77.Cdn, diags *diag.Diagnostics) Model {
	var stream *ModelStream
	if cdn.Stream != nil {
		stream = &ModelStream{
			OriginUrl: types.StringValue(cdn.Stream.OriginUrl),
			Password:  types.StringPointerValue(cdn.Stream.Password),
			QueryKey:  types.StringValue(cdn.Stream.QueryKey),
			Protocol:  types.StringValue(cdn.Stream.Protocol),
			Port:      types.Int64Value(int64(cdn.Stream.Port)),
			Path:      types.StringPointerValue(cdn.Stream.Path),
		}
	}

	cnames := r.getCnamesSet(ctx, cdn, diags)

	geoProtectionCountries := types.SetNull(types.StringType)
	if cdn.GeoProtection.Countries != nil {
		geoProtectionCountries = util.SetValueFrom(ctx, diags, types.StringType, *cdn.GeoProtection.Countries)
	}

	hotlinkProtectionDomains := types.SetNull(types.StringType)
	if cdn.HotlinkProtection.Domains != nil {
		hotlinkProtectionDomains = util.SetValueFrom(ctx, diags, types.StringType, *cdn.HotlinkProtection.Domains)
	}

	ipProtectionIps := types.SetNull(types.StringType)
	if cdn.IpProtection.Ips != nil {
		ipProtectionIps = util.SetValueFrom(ctx, diags, types.StringType, *cdn.IpProtection.Ips)
	}

	originHeaders := types.MapNull(types.StringType)
	if cdn.OriginHeaders != nil && !cdn.OriginHeaders.Custom.IsNull() && cdn.OriginHeaders.Custom.IsSpecified() {
		originHeaders = util.MapValueFrom(ctx, diags, types.StringType, cdn.OriginHeaders.Custom.MustGet())
	}

	queryStringParameters := types.SetNull(types.StringType)
	if cdn.QueryString.Parameters != nil {
		queryStringParameters = util.SetValueFrom(ctx, diags, types.StringType, *cdn.QueryString.Parameters)
	}

	conditionalFeatures := r.readConditionalFeatures(ctx, model, cdn, diags)

	if diags.HasError() {
		return model
	}

	return Model{
		Id:           model.Id,
		Label:        types.StringValue(cdn.Label),
		OriginId:     util.NullableToStringValue(cdn.OriginId),
		CreationTime: types.StringValue(cdn.CreationTime.Format(time.DateTime)),
		Url:          types.StringValue(cdn.Url),
		Stream:       stream,
		Cache: &ModelCache{
			MaxAge:                     util.IntPointerToInt64Value(cdn.Cache.MaxAge),
			MaxAge404:                  util.NullableIntToInt64Value(cdn.Cache.MaxAge404),
			RequestsWithCookiesEnabled: types.BoolPointerValue(cdn.Cache.RequestsWithCookiesEnabled),
		},
		Cnames: cnames,
		GeoProtection: &ModelGeoProtection{
			Countries: geoProtectionCountries,
			Type:      types.StringValue(string(cdn.GeoProtection.Type)),
		},
		Headers: &ModelHeaders{
			CorsEnabled:                 types.BoolPointerValue(cdn.Headers.CorsEnabled),
			CorsTimingEnabled:           types.BoolPointerValue(cdn.Headers.CorsTimingEnabled),
			CorsWildcardEnabled:         types.BoolPointerValue(cdn.Headers.CorsWildcardEnabled),
			HostHeaderForwardingEnabled: types.BoolPointerValue(cdn.Headers.HostHeaderForwardingEnabled),
			ContentDispositionType:      types.StringValue(string(*cdn.Headers.ContentDisposition.Type)),
		},
		HotlinkProtection: &ModelHotlinkProtection{
			Domains:            hotlinkProtectionDomains,
			Type:               types.StringValue(string(cdn.HotlinkProtection.Type)),
			EmptyRefererDenied: types.BoolValue(cdn.HotlinkProtection.EmptyRefererDenied),
		},
		HttpsRedirect: &ModelHttpsRedirect{
			Code:    util.IntPointerToInt64Value(cdn.HttpsRedirect.Code),
			Enabled: types.BoolValue(cdn.HttpsRedirect.Enabled),
		},
		IpProtection: &ModelIpProtection{
			Ips:  ipProtectionIps,
			Type: types.StringValue(string(cdn.IpProtection.Type)),
		},
		Mp4PseudoStreamingEnabled: types.BoolPointerValue(cdn.Mp4PseudoStreaming.Enabled),
		Note:                      util.NullableToStringValue(cdn.Note),
		OriginHeaders:             originHeaders,
		QueryString: &ModelQueryString{
			Parameters: queryStringParameters,
			IgnoreType: types.StringValue(string(cdn.QueryString.IgnoreType)),
		},
		RateLimitEnabled: types.BoolValue(cdn.RateLimit.Enabled),
		SecureToken: &ModelSecureToken{
			Token: types.StringPointerValue(cdn.SecureToken.Token),
			Type:  types.StringValue(string(cdn.SecureToken.Type)),
		},
		Ssl: &ModelSsl{
			Type:  types.StringValue(string(cdn.Ssl.Type)),
			SslId: types.StringPointerValue(cdn.Ssl.SslId),
		},
		ConditionalFeatures: conditionalFeatures,
	}
}

func (*Reader) getCnamesSet(ctx context.Context, cdn *cdn77.Cdn, diags *diag.Diagnostics) types.Set {
	cnames := make([]string, len(cdn.Cnames))

	for i, c := range cdn.Cnames {
		cnames[i] = c.Cname
	}

	return util.SetValueFrom(ctx, diags, types.StringType, cnames)
}

func (r *Reader) readConditionalFeatures(
	ctx context.Context,
	state Model,
	c *cdn77.Cdn,
	diags *diag.Diagnostics,
) *ModelConditionalFeatures {
	if c.ConditionalFeatures == nil {
		return nil
	}

	src := c.ConditionalFeatures

	configuration := r.readConditionalFeaturesConfiguration(diags, src)
	secrets := r.readConditionalFeaturesSecrets(ctx, diags, src, state.ConditionalFeatures)

	if diags.HasError() {
		return nil
	}

	isConfigEmpty := configuration.IsNull()
	isSecretsEmpty := secrets.IsNull() || len(secrets.Elements()) == 0

	if isConfigEmpty && isSecretsEmpty {
		return nil
	}

	return &ModelConditionalFeatures{
		Configuration: configuration,
		Secrets:       secrets,
	}
}

func (*Reader) readConditionalFeaturesConfiguration(
	diags *diag.Diagnostics,
	src *cdn77.ConditionalFeatures,
) types.String {
	if src.Configuration == nil || len(*src.Configuration) == 0 {
		return types.StringNull()
	}

	raw, err := json.Marshal(src.Configuration)
	if err != nil {
		diags.AddError("Failed to marshal conditional_features.configuration", err.Error())

		return types.StringNull()
	}

	raw, err = normalizeConditionalFeaturesEmptyConfigArray(raw)
	if err != nil {
		diags.AddError("Failed to normalize conditional_features.configuration", err.Error())

		return types.StringNull()
	}

	canon, err := canonicalizeJSON(string(raw))
	if err != nil {
		diags.AddError("Failed to canonicalize conditional_features.configuration", err.Error())

		return types.StringNull()
	}

	return types.StringValue(canon)
}

func (*Reader) readConditionalFeaturesSecrets(
	ctx context.Context,
	diags *diag.Diagnostics,
	src *cdn77.ConditionalFeatures,
	state *ModelConditionalFeatures,
) types.Map {
	if src.Secrets != nil && len(*src.Secrets) > 0 {
		secrets, diagnostics := types.MapValueFrom(ctx, types.StringType, *src.Secrets)
		diags.Append(diagnostics...)

		return secrets
	}

	if state != nil &&
		!state.Secrets.IsNull() &&
		!state.Secrets.IsUnknown() {
		return state.Secrets
	}

	return types.MapValueMust(types.StringType, map[string]attr.Value{})
}

func normalizeConditionalFeaturesEmptyConfigArray(raw []byte) ([]byte, error) { //nolint:cyclop
	var rules []map[string]any
	if err := json.Unmarshal(raw, &rules); err != nil {
		return nil, err
	}

	for _, rule := range rules {
		thenValue, exists := rule["then"]
		if !exists {
			continue
		}

		thenActions, ok := thenValue.([]any)
		if !ok {
			continue
		}

		for _, actionValue := range thenActions {
			actionMap, ok := actionValue.(map[string]any)
			if !ok {
				continue
			}

			configValue, exists := actionMap["config"]
			if !exists {
				continue
			}

			if configArray, ok := configValue.([]any); ok && len(configArray) == 0 {
				actionMap["config"] = map[string]any{}
			}
		}
	}

	normalizedJSON, err := json.Marshal(rules)
	if err != nil {
		return nil, err
	}

	normalizedJSON = bytes.TrimSpace(normalizedJSON)

	return normalizedJSON, nil
}
