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

type CdnModel struct {
	Id                        types.Int64                `tfsdk:"id"`
	Cnames                    types.Set                  `tfsdk:"cnames"`
	CreationTime              types.String               `tfsdk:"creation_time"`
	Label                     types.String               `tfsdk:"label"`
	Note                      types.String               `tfsdk:"note"`
	OriginId                  types.String               `tfsdk:"origin_id"`
	OriginProtectionEnabled   types.Bool                 `tfsdk:"origin_protection_enabled"`
	Url                       types.String               `tfsdk:"url"`
	Cache                     *CdnModelCache             `tfsdk:"cache"`
	SecureToken               *CdnModelSecureToken       `tfsdk:"secure_token"`
	QueryString               *CdnModelQueryString       `tfsdk:"query_string"`
	Headers                   *CdnModelHeaders           `tfsdk:"headers"`
	HttpsRedirect             *CdnModelHttpsRedirect     `tfsdk:"https_redirect"`
	Mp4PseudoStreamingEnabled types.Bool                 `tfsdk:"mp4_pseudo_streaming_enabled"`
	WafEnabled                types.Bool                 `tfsdk:"waf_enabled"`
	Ssl                       *CdnModelSsl               `tfsdk:"ssl"`
	Stream                    *CdnModelStream            `tfsdk:"stream"`
	HotlinkProtection         *CdnModelHotlinkProtection `tfsdk:"hotlink_protection"`
	IpProtection              *CdnModelIpProtection      `tfsdk:"ip_protection"`
	GeoProtection             *CdnModelGeoProtection     `tfsdk:"geo_protection"`
	RateLimitEnabled          types.Bool                 `tfsdk:"rate_limit_enabled"`
	OriginHeaders             types.Map                  `tfsdk:"origin_headers"`
}

type CdnModelCache struct {
	MaxAge                     types.Int64 `tfsdk:"max_age"`
	MaxAge404                  types.Int64 `tfsdk:"max_age_404"`
	RequestsWithCookiesEnabled types.Bool  `tfsdk:"requests_with_cookies_enabled"`
}

type CdnModelSecureToken struct {
	Token types.String `tfsdk:"token"`
	Type  types.String `tfsdk:"type"`
}

type CdnModelQueryString struct {
	Parameters types.Set    `tfsdk:"parameters"`
	IgnoreType types.String `tfsdk:"ignore_type"`
}

type CdnModelHeaders struct {
	CorsEnabled                 types.Bool   `tfsdk:"cors_enabled"`
	CorsTimingEnabled           types.Bool   `tfsdk:"cors_timing_enabled"`
	CorsWildcardEnabled         types.Bool   `tfsdk:"cors_wildcard_enabled"`
	HostHeaderForwardingEnabled types.Bool   `tfsdk:"host_header_forwarding_enabled"`
	ContentDispositionType      types.String `tfsdk:"content_disposition_type"`
}

type CdnModelHttpsRedirect struct {
	Code    types.Int64 `tfsdk:"code"`
	Enabled types.Bool  `tfsdk:"enabled"`
}

type CdnModelSsl struct {
	Type  types.String `tfsdk:"type"`
	SslId types.String `tfsdk:"ssl_id"`
}

type CdnModelStream struct {
	OriginUrl types.String `tfsdk:"origin_url"`
	Password  types.String `tfsdk:"password"`
	QueryKey  types.String `tfsdk:"query_key"`
	Protocol  types.String `tfsdk:"protocol"`
	Port      types.Int64  `tfsdk:"port"`
	Path      types.String `tfsdk:"path"`
}

type CdnModelHotlinkProtection struct {
	Domains            types.Set    `tfsdk:"domains"`
	Type               types.String `tfsdk:"type"`
	EmptyRefererDenied types.Bool   `tfsdk:"empty_referer_denied"`
}

type CdnModelIpProtection struct {
	Ips  types.Set    `tfsdk:"ips"`
	Type types.String `tfsdk:"type"`
}

type CdnModelGeoProtection struct {
	Countries types.Set    `tfsdk:"countries"`
	Type      types.String `tfsdk:"type"`
}

type CdnDataReader struct {
	ctx                   context.Context
	client                cdn77.ClientWithResponsesInterface
	removeMissingResource bool
}

func NewCdnDataSourceReader(ctx context.Context, client cdn77.ClientWithResponsesInterface) *CdnDataReader {
	return &CdnDataReader{ctx: ctx, client: client, removeMissingResource: false}
}

func NewCdnResourceReader(ctx context.Context, client cdn77.ClientWithResponsesInterface) *CdnDataReader {
	return &CdnDataReader{ctx: ctx, client: client, removeMissingResource: true}
}

func (d *CdnDataReader) Read(provider StateProvider, diags *diag.Diagnostics, state *tfsdk.State) { //nolint:cyclop
	var data CdnModel
	if diags.Append(provider.Get(d.ctx, &data)...); diags.HasError() {
		return
	}

	const errMessage = "Failed to fetch CDN"

	response, err := d.client.CdnDetailWithResponse(d.ctx, int(data.Id.ValueInt64()))
	if err != nil {
		diags.AddError(errMessage, err.Error())

		return
	}

	if d.removeMissingResource &&
		maybeRemoveMissingResource(d.ctx, response.StatusCode(), data.Id.ValueInt64(), state) {
		return
	}

	if !util.CheckResponse(diags, errMessage, response, response.JSON404, response.JSONDefault) {
		return
	}

	cdn := *response.JSON200

	cnamesRaw := make([]string, len(cdn.Cnames))

	for i, c := range cdn.Cnames {
		cnamesRaw[i] = c.Cname
	}

	cnames, ds := types.SetValueFrom(d.ctx, types.StringType, cnamesRaw)
	if ds != nil {
		diags.Append(ds...)

		return
	}

	queryStringParameters := types.SetNull(types.StringType)
	if cdn.QueryString.Parameters != nil {
		queryStringParameters, ds = types.SetValueFrom(d.ctx, types.StringType, *cdn.QueryString.Parameters)
		if ds != nil {
			diags.Append(ds...)

			return
		}
	}

	var stream *CdnModelStream
	if cdn.Stream != nil {
		stream = &CdnModelStream{
			OriginUrl: types.StringValue(cdn.Stream.OriginUrl),
			Password:  types.StringPointerValue(cdn.Stream.Password),
			QueryKey:  types.StringValue(cdn.Stream.QueryKey),
			Protocol:  types.StringValue(cdn.Stream.Protocol),
			Port:      types.Int64Value(int64(cdn.Stream.Port)),
			Path:      types.StringPointerValue(cdn.Stream.Path),
		}
	}

	hotlinkProtectionDomains := types.SetNull(types.StringType)
	if cdn.HotlinkProtection.Domains != nil {
		hotlinkProtectionDomains, ds = types.SetValueFrom(d.ctx, types.StringType, *cdn.HotlinkProtection.Domains)
		if ds != nil {
			diags.Append(ds...)

			return
		}
	}

	ipProtectionIps := types.SetNull(types.StringType)
	if cdn.IpProtection.Ips != nil {
		ipProtectionIps, ds = types.SetValueFrom(d.ctx, types.StringType, *cdn.IpProtection.Ips)
		if ds != nil {
			diags.Append(ds...)

			return
		}
	}

	geoProtectionCountries := types.SetNull(types.StringType)
	if cdn.GeoProtection.Countries != nil {
		geoProtectionCountries, ds = types.SetValueFrom(d.ctx, types.StringType, *cdn.GeoProtection.Countries)
		if ds != nil {
			diags.Append(ds...)

			return
		}
	}

	originHeaders := types.MapNull(types.StringType)
	if cdn.OriginHeaders != nil && !cdn.OriginHeaders.Custom.IsNull() && cdn.OriginHeaders.Custom.IsSpecified() {
		originHeaders, ds = types.MapValueFrom(d.ctx, types.StringType, cdn.OriginHeaders.Custom.MustGet())
		if ds != nil {
			diags.Append(ds...)

			return
		}
	}

	data = CdnModel{
		Id:                      data.Id,
		Cnames:                  cnames,
		CreationTime:            types.StringValue(cdn.CreationTime.Format(time.DateTime)),
		Label:                   types.StringValue(cdn.Label),
		Note:                    util.NullableToStringValue(cdn.Note),
		OriginId:                util.NullableToStringValue(cdn.OriginId),
		OriginProtectionEnabled: types.BoolValue(cdn.OriginProtection.Enabled),
		Url:                     types.StringValue(cdn.Url),
		Cache: &CdnModelCache{
			MaxAge:                     util.IntPointerToInt64Value(cdn.Cache.MaxAge),
			MaxAge404:                  util.NullableIntToInt64Value(cdn.Cache.MaxAge404),
			RequestsWithCookiesEnabled: types.BoolPointerValue(cdn.Cache.RequestsWithCookiesEnabled),
		},
		SecureToken: &CdnModelSecureToken{
			Token: types.StringPointerValue(cdn.SecureToken.Token),
			Type:  types.StringValue(string(cdn.SecureToken.Type)),
		},
		QueryString: &CdnModelQueryString{
			Parameters: queryStringParameters,
			IgnoreType: types.StringValue(string(cdn.QueryString.IgnoreType)),
		},
		Headers: &CdnModelHeaders{
			CorsEnabled:                 types.BoolPointerValue(cdn.Headers.CorsEnabled),
			CorsTimingEnabled:           types.BoolPointerValue(cdn.Headers.CorsTimingEnabled),
			CorsWildcardEnabled:         types.BoolPointerValue(cdn.Headers.CorsWildcardEnabled),
			HostHeaderForwardingEnabled: types.BoolPointerValue(cdn.Headers.HostHeaderForwardingEnabled),
			ContentDispositionType:      types.StringValue(string(*cdn.Headers.ContentDisposition.Type)),
		},
		HttpsRedirect: &CdnModelHttpsRedirect{
			Code:    util.IntPointerToInt64Value(cdn.HttpsRedirect.Code),
			Enabled: types.BoolValue(cdn.HttpsRedirect.Enabled),
		},
		Mp4PseudoStreamingEnabled: types.BoolPointerValue(cdn.Mp4PseudoStreaming.Enabled),
		WafEnabled:                types.BoolValue(cdn.Waf.Enabled),
		Ssl: &CdnModelSsl{
			Type:  types.StringValue(string(cdn.Ssl.Type)),
			SslId: types.StringPointerValue(cdn.Ssl.SslId),
		},
		Stream: stream,
		HotlinkProtection: &CdnModelHotlinkProtection{
			Domains:            hotlinkProtectionDomains,
			Type:               types.StringValue(string(cdn.HotlinkProtection.Type)),
			EmptyRefererDenied: types.BoolValue(cdn.HotlinkProtection.EmptyRefererDenied),
		},
		IpProtection: &CdnModelIpProtection{
			Ips:  ipProtectionIps,
			Type: types.StringValue(string(cdn.IpProtection.Type)),
		},
		GeoProtection: &CdnModelGeoProtection{
			Countries: geoProtectionCountries,
			Type:      types.StringValue(string(cdn.GeoProtection.Type)),
		},
		RateLimitEnabled: types.BoolValue(cdn.RateLimit.Enabled),
		OriginHeaders:    originHeaders,
	}

	diags.Append(state.Set(d.ctx, &data)...)
}
