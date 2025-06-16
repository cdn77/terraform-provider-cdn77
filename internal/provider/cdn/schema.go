package cdn

import (
	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type Model struct {
	Id                        types.Int64             `tfsdk:"id"`
	Label                     types.String            `tfsdk:"label"`
	OriginId                  types.String            `tfsdk:"origin_id"`
	CreationTime              types.String            `tfsdk:"creation_time"`
	Url                       types.String            `tfsdk:"url"`
	Stream                    *ModelStream            `tfsdk:"stream"`
	Cache                     *ModelCache             `tfsdk:"cache"`
	Cnames                    types.Set               `tfsdk:"cnames"`
	GeoProtection             *ModelGeoProtection     `tfsdk:"geo_protection"`
	Headers                   *ModelHeaders           `tfsdk:"headers"`
	HotlinkProtection         *ModelHotlinkProtection `tfsdk:"hotlink_protection"`
	HttpsRedirect             *ModelHttpsRedirect     `tfsdk:"https_redirect"`
	IpProtection              *ModelIpProtection      `tfsdk:"ip_protection"`
	Mp4PseudoStreamingEnabled types.Bool              `tfsdk:"mp4_pseudo_streaming_enabled"`
	Note                      types.String            `tfsdk:"note"`
	OriginHeaders             types.Map               `tfsdk:"origin_headers"`
	QueryString               *ModelQueryString       `tfsdk:"query_string"`
	RateLimitEnabled          types.Bool              `tfsdk:"rate_limit_enabled"`
	SecureToken               *ModelSecureToken       `tfsdk:"secure_token"`
	Ssl                       *ModelSsl               `tfsdk:"ssl"`
}

type ModelStream struct {
	OriginUrl types.String `tfsdk:"origin_url"`
	Password  types.String `tfsdk:"password"`
	QueryKey  types.String `tfsdk:"query_key"`
	Protocol  types.String `tfsdk:"protocol"`
	Port      types.Int64  `tfsdk:"port"`
	Path      types.String `tfsdk:"path"`
}

type ModelCache struct {
	MaxAge                     types.Int64 `tfsdk:"max_age"`
	MaxAge404                  types.Int64 `tfsdk:"max_age_404"`
	RequestsWithCookiesEnabled types.Bool  `tfsdk:"requests_with_cookies_enabled"`
}

type ModelGeoProtection struct {
	Countries types.Set    `tfsdk:"countries"`
	Type      types.String `tfsdk:"type"`
}

type ModelHeaders struct {
	CorsEnabled                 types.Bool   `tfsdk:"cors_enabled"`
	CorsTimingEnabled           types.Bool   `tfsdk:"cors_timing_enabled"`
	CorsWildcardEnabled         types.Bool   `tfsdk:"cors_wildcard_enabled"`
	HostHeaderForwardingEnabled types.Bool   `tfsdk:"host_header_forwarding_enabled"`
	ContentDispositionType      types.String `tfsdk:"content_disposition_type"`
}

type ModelHotlinkProtection struct {
	Domains            types.Set    `tfsdk:"domains"`
	Type               types.String `tfsdk:"type"`
	EmptyRefererDenied types.Bool   `tfsdk:"empty_referer_denied"`
}

type ModelHttpsRedirect struct {
	Code    types.Int64 `tfsdk:"code"`
	Enabled types.Bool  `tfsdk:"enabled"`
}

type ModelIpProtection struct {
	Ips  types.Set    `tfsdk:"ips"`
	Type types.String `tfsdk:"type"`
}

type ModelQueryString struct {
	Parameters types.Set    `tfsdk:"parameters"`
	IgnoreType types.String `tfsdk:"ignore_type"`
}

type ModelSecureToken struct {
	Token types.String `tfsdk:"token"`
	Type  types.String `tfsdk:"type"`
}

type ModelSsl struct {
	Type  types.String `tfsdk:"type"`
	SslId types.String `tfsdk:"ssl_id"`
}

func CreateResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CDN resource allows you to manage your CDNs",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:      true,
				Description:   "ID of the CDN. This is also used as the CDN URL",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"label": schema.StringAttribute{
				Description: "The label helps you to identify your CDN",
				Required:    true,
			},
			"origin_id": schema.StringAttribute{
				Description: "ID (UUID) of attached Origin (content source for CDN)",
				Required:    true,
			},
			"creation_time": schema.StringAttribute{
				Computed:      true,
				Description:   "Timestamp when CDN was created",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"url": schema.StringAttribute{
				Computed: true,
				Description: "URL of the CDN. Automatically generated when the CDN is created. " +
					"The number is the same as the CDN ID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"stream": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"origin_url": schema.StringAttribute{
						Optional: true,
					},
					"password": schema.StringAttribute{
						Optional:  true,
						Sensitive: true,
					},
					"query_key": schema.StringAttribute{
						Optional: true,
					},
					"protocol": schema.StringAttribute{
						Optional: true,
					},
					"port": schema.Int64Attribute{
						Optional: true,
					},
					"path": schema.StringAttribute{
						Optional: true,
					},
				},
				Optional:    true,
				Description: "Detail parameters of stream CDN",
			},
			"cache": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"max_age": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "In minutes",
						Validators: []validator.Int64{int64validator.OneOf(
							int64(cdn77.MaxAgeN10), int64(cdn77.MaxAgeN30), int64(cdn77.MaxAgeN60),
							int64(cdn77.MaxAgeN240), int64(cdn77.MaxAgeN720), int64(cdn77.MaxAgeN1440),
							int64(cdn77.MaxAgeN2160), int64(cdn77.MaxAgeN2880), int64(cdn77.MaxAgeN4320),
							int64(cdn77.MaxAgeN5760), int64(cdn77.MaxAgeN7200), int64(cdn77.MaxAgeN8640),
							int64(cdn77.MaxAgeN10800), int64(cdn77.MaxAgeN11520), int64(cdn77.MaxAgeN12960),
							int64(cdn77.MaxAgeN14400), int64(cdn77.MaxAgeN15840), int64(cdn77.MaxAgeN17280),
						)},
						Default: int64default.StaticInt64(int64(cdn77.MaxAgeN17280)),
					},
					"max_age_404": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "In seconds",
						Validators: []validator.Int64{int64validator.OneOf(
							int64(cdn77.MaxAge404N1), int64(cdn77.MaxAge404N5), int64(cdn77.MaxAge404N10),
							int64(cdn77.MaxAge404N30), int64(cdn77.MaxAge404N60), int64(cdn77.MaxAge404N300),
							int64(cdn77.MaxAge404N3600),
						)},
						Default: util.Int64NullDefault(),
					},
					"requests_with_cookies_enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Description: "When disabled, requests with cookies will ignore changing cookie " +
							"headers allowing all requests to hit the cache. When enabled, requests with cookies " +
							"will be handled separately, so when the cookie changes it will not hit the previously " +
							"cached request with different cookie options.",
						Default: booldefault.StaticBool(true),
					},
				},
				Optional: true,
				Computed: true,
				Description: "Your files will remain cached for the specified duration, after which your " +
					"origin will be checked for an updated version of your files. Expiry/cache-control headers " +
					"override this setting.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"max_age":                       basetypes.Int64Type{},
						"max_age_404":                   basetypes.Int64Type{},
						"requests_with_cookies_enabled": basetypes.BoolType{},
					},
					map[string]attr.Value{
						"max_age":                       types.Int64Value(int64(cdn77.MaxAgeN17280)),
						"max_age_404":                   types.Int64Null(),
						"requests_with_cookies_enabled": types.BoolValue(true),
					},
				)),
			},
			"cnames": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "CNAME assigned to CDN. " +
					"CNAME should be mapped via DNS to CDN URL. " +
					"Otherwise it's not possible to generate an SSL certificate for any related CNAME.",
				Default: setdefault.StaticValue(types.SetValueMust(types.StringType, nil)),
			},
			"geo_protection": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"countries": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "We are using ISO 3166-1 alpha-2 code",
					},
					"type": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Description: `With "type": "blocklist" all countries set in the "countries" ` +
							`parameter are not allowed. With "type": "passlist" only countries set in the ` +
							`"countries" parameter are allowed.`,
						Validators: []validator.String{stringvalidator.OneOf(
							string(cdn77.Blocklist),
							string(cdn77.Disabled),
							string(cdn77.Passlist),
						)},
						Default: stringdefault.StaticString(string(cdn77.Disabled)),
					},
				},
				Optional: true,
				Computed: true,
				Description: "Geo protection enables you to control which countries can access your " +
					"content directly",
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"countries": basetypes.SetType{ElemType: basetypes.StringType{}},
						"type":      basetypes.StringType{},
					},
					map[string]attr.Value{
						"countries": types.SetNull(basetypes.StringType{}),
						"type":      types.StringValue(string(cdn77.Disabled)),
					},
				)),
			},
			"headers": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"cors_enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Description: `The "Access-Control-Allow-Origin:" response header will always act in ` +
							`accordance with the "Origin:" request header sent by the client. For example, a ` +
							`request including the HTTP header "Origin: https://www.cdn77.com" will translate to ` +
							`the response header "Access-Control-Allow-Origin: https://www.cdn77.com". Files remain ` +
							`cached while the request/response header changes.`,
						Default: booldefault.StaticBool(false),
					},
					"cors_timing_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: `When enabled "Timing-Allow-Origin" CORS header is set`,
						Default:     booldefault.StaticBool(false),
					},
					"cors_wildcard_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: `When enabled the wildcard value (*) is set in CORS headers`,
						Default:     booldefault.StaticBool(false),
					},
					"host_header_forwarding_enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Description: "When fetching the content from the origin server, our edge servers " +
							"will pass the host header that was included in the request between the user " +
							"and our edge server.",
						Default: booldefault.StaticBool(false),
					},
					"content_disposition_type": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Description: `When the "type" is set to "parameter" the Content-Disposition is ` +
							`defined by the "cd" parameter in the URL, often set to "attachment". The filename is ` +
							`specified using the "fn" parameter in the URL.`,
						Validators: []validator.String{stringvalidator.OneOf(
							string(cdn77.ContentDispositionTypeNone),
							string(cdn77.ContentDispositionTypeParameter),
						)},
						Default: stringdefault.StaticString(string(cdn77.ContentDispositionTypeNone)),
					},
				},
				Optional: true,
				Computed: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"cors_enabled":                   basetypes.BoolType{},
						"cors_timing_enabled":            basetypes.BoolType{},
						"cors_wildcard_enabled":          basetypes.BoolType{},
						"host_header_forwarding_enabled": basetypes.BoolType{},
						"content_disposition_type":       basetypes.StringType{},
					},
					map[string]attr.Value{
						"cors_enabled":                   types.BoolValue(false),
						"cors_timing_enabled":            types.BoolValue(false),
						"cors_wildcard_enabled":          types.BoolValue(false),
						"host_header_forwarding_enabled": types.BoolValue(false),
						"content_disposition_type": types.StringValue(
							string(cdn77.ContentDispositionTypeNone),
						),
					},
				)),
			},
			"hotlink_protection": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"domains": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"type": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Description: `With "type": "blocklist" all domains set in the "domains" parameter ` +
							`are not allowed. With "type": "passlist" only domains in "domains" parameter are allowed.`,
						Validators: []validator.String{stringvalidator.OneOf(
							string(cdn77.Blocklist),
							string(cdn77.Disabled),
							string(cdn77.Passlist),
						)},
						Default: stringdefault.StaticString(string(cdn77.Disabled)),
					},
					"empty_referer_denied": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Description: "Enabling this parameter prevents your content from being directly " +
							"accessed by sources that send empty referrers.",
						Default: booldefault.StaticBool(false),
					},
				},
				Optional: true,
				Computed: true,
				Description: "Hotlink protection enables you to control which hostnames/domains can link to " +
					"and access your content directly",
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"domains":              basetypes.SetType{ElemType: basetypes.StringType{}},
						"type":                 basetypes.StringType{},
						"empty_referer_denied": basetypes.BoolType{},
					},
					map[string]attr.Value{
						"domains":              types.SetNull(basetypes.StringType{}),
						"type":                 types.StringValue(string(cdn77.Disabled)),
						"empty_referer_denied": types.BoolValue(false),
					},
				)),
			},
			"https_redirect": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"code": schema.Int64Attribute{
						Optional: true,
						Description: "301 for permanent redirect and 302 for temporary redirect. " +
							"If you are not sure, select the default 301 code.",
						Validators: []validator.Int64{int64validator.OneOf(301, 302)},
					},
					"enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
				},
				Optional: true,
				Computed: true,
				Description: "If enabled, all requests via HTTP are redirected to HTTPS. " +
					"Verify HTTPS availability of CNAMEs before activating, if applicable.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"code":    basetypes.Int64Type{},
						"enabled": basetypes.BoolType{},
					},
					map[string]attr.Value{
						"code":    types.Int64Null(),
						"enabled": types.BoolValue(false),
					},
				)),
			},
			"ip_protection": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"ips": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"type": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Description: `With "type": "blocklist" all IP addresses set in the "ips" parameter are not ` +
							`allowed. With "type": "passlist" only IP addresses in "ips" parameter are allowed.`,
						Validators: []validator.String{stringvalidator.OneOf(
							string(cdn77.Blocklist),
							string(cdn77.Disabled),
							string(cdn77.Passlist),
						)},
						Default: stringdefault.StaticString(string(cdn77.Disabled)),
					},
				},
				Optional: true,
				Computed: true,
				Description: "IP protection enables you to control which networks can access your " +
					"content directly",
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"ips":  basetypes.SetType{ElemType: basetypes.StringType{}},
						"type": basetypes.StringType{},
					},
					map[string]attr.Value{
						"ips":  types.SetNull(basetypes.StringType{}),
						"type": types.StringValue(string(cdn77.Disabled)),
					},
				)),
			},
			"mp4_pseudo_streaming_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Description: "Turn this option on if using a flash-based video player with MP4 files. " +
					"Pseudo-streaming is used mainly in flash players. HTML5 players use range-requests. " +
					`When enabled the "query_string" option must be set to ignore all parameters.`,
				Default: booldefault.StaticBool(false),
			},
			"note": schema.StringAttribute{
				Description: "Optional note",
				Optional:    true,
			},
			"origin_headers": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Custom HTTP headers included in requests sent to the origin server",
			},
			"query_string": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"parameters": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: `List of parameters used when "ignore_type" is set to "list"`,
					},
					"ignore_type": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Validators: []validator.String{stringvalidator.OneOf(
							string(cdn77.QueryStringIgnoreTypeAll),
							string(cdn77.QueryStringIgnoreTypeList),
							string(cdn77.QueryStringIgnoreTypeNone),
						)},
						Default: stringdefault.StaticString(string(cdn77.QueryStringIgnoreTypeNone)),
					},
				},
				Optional: true,
				Computed: true,
				Description: "Enabling this feature will ignore the query string, allowing URLs with " +
					"query strings to cache properly. This is particularly useful if you tag your URLs with " +
					"tracking/marketing parameters, for example.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"parameters":  basetypes.SetType{ElemType: basetypes.StringType{}},
						"ignore_type": basetypes.StringType{},
					},
					map[string]attr.Value{
						"parameters":  types.SetNull(basetypes.StringType{}),
						"ignore_type": types.StringValue(string(cdn77.QueryStringIgnoreTypeNone)),
					},
				)),
			},
			"rate_limit_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Description: `When enabled, this feature limits the data transfer rate by setting ` +
					`"limit_rate" based on the "rs" URL parameter and "limit_rate_after" by the value from the "ri" ` +
					`URL parameter.`,
				Default: booldefault.StaticBool(false),
			},
			"secure_token": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"token": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Token length is between 8 and 64 characters.",
						Validators:  []validator.String{stringvalidator.LengthBetween(8, 64)},
					},
					"type": schema.StringAttribute{
						Optional: true,
						Computed: true,
						MarkdownDescription: `<ul>
	<li>parameter - Token will be in the query string - e.g.: /video.mp4?secure=MY_SECURE_TOKEN.</li>
	<li>path - Token will be in the path - e.g.: /MY_SECURE_TOKEN/video.mp4.</li>
	<li>none - Use to disable secure token.</li>
	<li>highwinds</li>
</ul>`,
						Validators: []validator.String{stringvalidator.OneOf(
							string(cdn77.SecureTokenTypeHighwinds),
							string(cdn77.SecureTokenTypeNone),
							string(cdn77.SecureTokenTypeParameter),
							string(cdn77.SecureTokenTypePath),
						)},
						Default: stringdefault.StaticString(string(cdn77.SecureTokenTypeNone)),
					},
				},
				Optional: true,
				Computed: true,
				Description: "This feature allows you to serve your content using signed URLs. " +
					"You can enable your users to download secured content from the CDN with a valid hash. " +
					"Note: When you check this option, make sure to generate secured links to access your content.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{"token": basetypes.StringType{}, "type": basetypes.StringType{}},
					map[string]attr.Value{
						"token": types.StringNull(),
						"type":  types.StringValue(string(cdn77.SecureTokenTypeNone)),
					},
				)),
			},
			"ssl": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Possible values: instantSsl, none, SNI",
						Validators: []validator.String{stringvalidator.OneOf(
							string(cdn77.InstantSsl),
							string(cdn77.None),
							string(cdn77.SNI),
						)},
						Default: stringdefault.StaticString(string(cdn77.InstantSsl)),
					},
					"ssl_id": schema.StringAttribute{
						Optional:    true,
						Description: "ID (UUID) of the SSL certificate",
					},
				},
				Optional: true,
				Computed: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"type":   basetypes.StringType{},
						"ssl_id": basetypes.StringType{},
					},
					map[string]attr.Value{
						"type":   types.StringValue(string(cdn77.InstantSsl)),
						"ssl_id": types.StringNull(),
					},
				)),
			},
		},
	}
}
