package shared

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/oapi-codegen/nullable"
)

var urlBasePathRegexp = regexp.MustCompile(`^[a-zA-Z0-9.\-_\[\]+%*](/?[a-zA-Z0-9.\-_\[\]+%*])*$`)

type UrlModel struct {
	Url      types.String `tfsdk:"url"`
	UrlParts types.Object `tfsdk:"url_parts"`
}

func (m UrlModel) Parts(ctx context.Context) (
	scheme string,
	host string,
	port nullable.Nullable[int],
	basePath nullable.Nullable[string],
) {
	var model UrlPartsModel
	if diags := m.UrlParts.As(ctx, &model, basetypes.ObjectAsOptions{}); diags != nil {
		panic(fmt.Sprintf("Failed to convert types.Object to UrlPartsModel: %v", diags))
	}

	return model.Scheme.ValueString(),
		model.Host.ValueString(),
		util.Int32ValueToNullable[int](model.Port),
		util.StringValueToNullable(model.BasePath)
}

type UrlPartsModel struct {
	Scheme   types.String `tfsdk:"scheme"`
	Host     types.String `tfsdk:"host"`
	Port     types.Int32  `tfsdk:"port"`
	BasePath types.String `tfsdk:"base_path"`
}

func NewUrlPartsModel(
	scheme string,
	host string,
	port nullable.Nullable[int],
	basePath nullable.Nullable[string],
) *UrlPartsModel {
	return &UrlPartsModel{
		Scheme:   types.StringValue(scheme),
		Host:     types.StringValue(host),
		Port:     util.NullableIntToInt32Value(port),
		BasePath: util.NullableToStringValue(basePath),
	}
}

func NewUrlModel(
	ctx context.Context,
	scheme string,
	host string,
	port nullable.Nullable[int],
	basePath nullable.Nullable[string],
) UrlModel {
	urlParts := NewUrlPartsModel(scheme, host, port, basePath)
	urlPartsTypes := map[string]attr.Type{
		"scheme":    urlParts.Scheme.Type(ctx),
		"host":      urlParts.Host.Type(ctx),
		"port":      urlParts.Port.Type(ctx),
		"base_path": urlParts.BasePath.Type(ctx),
	}

	urlPartsObject, diags := types.ObjectValueFrom(ctx, urlPartsTypes, urlParts)
	if diags != nil {
		panic(fmt.Sprintf("Failed to convert UrlPartsModel to types.Object: %v", diags))
	}

	return UrlModel{Url: urlFromParts(urlParts), UrlParts: urlPartsObject}
}

func WithUrlSchemaAttrs(s schema.Schema) schema.Schema {
	return withMaybeComputableUrlSchemaAttrs(s, false)
}

func WithComputedUrlSchemaAttrs(s schema.Schema) schema.Schema {
	return withMaybeComputableUrlSchemaAttrs(s, true)
}

func withMaybeComputableUrlSchemaAttrs(s schema.Schema, computed bool) schema.Schema {
	urlOrSourceValidator := stringvalidator.ExactlyOneOf(path.MatchRoot("url"), path.MatchRoot("url_parts"))
	s.Attributes["url"] = schema.StringAttribute{
		Description: `Absolute URL of this resource. Alternative to the attribute "url_parts".`,
		Optional:    util.If(computed, false, true),
		Computed:    true,
		Validators: util.If(
			computed,
			[]validator.String{UrlStringValidator{}},
			[]validator.String{UrlStringValidator{}, urlOrSourceValidator},
		),
		PlanModifiers: []planmodifier.String{UrlAndUrlPartsPlanModifier{}, stringplanmodifier.UseStateForUnknown()},
	}
	s.Attributes["url_parts"] = schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"scheme": schema.StringAttribute{
				Description: `URL scheme; can be either "http" or "https"`,
				Required:    util.If(computed, false, true),
				Computed:    computed,
				Validators:  []validator.String{stringvalidator.OneOf("http", "https")},
			},
			"host": schema.StringAttribute{
				Description: "Network host; can be a domain name or an IP address",
				Required:    util.If(computed, false, true),
				Computed:    computed,
			},
			"port": schema.Int32Attribute{
				Description: "Port number between 1 and 65535 (if not specified, default scheme port is used)",
				Optional:    true,
				Validators:  []validator.Int32{int32validator.Between(1, 65535)},
			},
			"base_path": schema.StringAttribute{
				Description: "Path to the directory where the content is stored",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
					stringvalidator.RegexMatches(
						urlBasePathRegexp,
						"mustn't begin or end with a slash "+
							"and must consist only of alphanumeric and following characters: .-_[]+%*/",
					),
				},
			},
		},
		Optional:      util.If(computed, false, true),
		Computed:      true,
		Description:   "Set of attributes describing the resource URL. Alternative to the attribute \"url\".",
		Validators:    util.If(computed, nil, []validator.Object{urlOrSourceValidator.(validator.Object)}),
		PlanModifiers: []planmodifier.Object{UrlAndUrlPartsPlanModifier{}, objectplanmodifier.UseStateForUnknown()},
	}

	return s
}

type UrlAndUrlPartsPlanModifier struct{}

func (UrlAndUrlPartsPlanModifier) Description(context.Context) string {
	return "sets the URL if URL parts are set and vice versa"
}

func (m UrlAndUrlPartsPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (UrlAndUrlPartsPlanModifier) PlanModifyString(
	ctx context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	if !req.PlanValue.IsUnknown() {
		return
	}

	diags := resp.Diagnostics
	var urlPartsObj types.Object
	partsPath := req.Path.ParentPath().AtName("url_parts")

	if diags.Append(req.Plan.GetAttribute(ctx, partsPath, &urlPartsObj)...); diags.HasError() {
		return
	}

	if urlPartsObj.IsUnknown() || urlPartsObj.IsNull() {
		return
	}

	var urlParts UrlPartsModel
	if diags.Append(urlPartsObj.As(ctx, &urlParts, basetypes.ObjectAsOptions{})...); diags.HasError() {
		return
	}

	resp.PlanValue = urlFromParts(&urlParts)
}

func (UrlAndUrlPartsPlanModifier) PlanModifyObject(
	ctx context.Context,
	req planmodifier.ObjectRequest,
	resp *planmodifier.ObjectResponse,
) {
	if !req.PlanValue.IsUnknown() {
		return
	}

	diags := resp.Diagnostics
	var urlValue types.String

	if diags.Append(req.Plan.GetAttribute(ctx, req.Path.ParentPath().AtName("url"), &urlValue)...); diags.HasError() {
		return
	}

	if urlValue.IsUnknown() {
		return
	}

	u, err := url.Parse(urlValue.ValueString())
	if err != nil {
		diags.AddError("Invalid URL", fmt.Sprintf("Failed to parse URL: %s", err))

		return
	}

	port := nullable.NewNullNullable[int]()
	if i, err := strconv.Atoi(u.Port()); err == nil {
		port.Set(i)
	}

	u.Path = strings.TrimLeft(u.Path, "/")
	basePath := util.If(u.Path == "", nullable.NewNullNullable[string](), nullable.NewNullableWithValue[string](u.Path))
	urlParts := NewUrlPartsModel(u.Scheme, u.Hostname(), port, basePath)

	urlPartsTypes := map[string]attr.Type{
		"scheme":    types.StringType,
		"host":      types.StringType,
		"port":      types.Int32Type,
		"base_path": types.StringType,
	}

	urlPartsObject, objDiags := types.ObjectValueFrom(ctx, urlPartsTypes, urlParts)
	if objDiags.HasError() {
		diags.Append(objDiags...)

		return
	}

	resp.PlanValue = urlPartsObject
}

type UrlStringValidator struct{}

func (UrlStringValidator) Description(context.Context) string {
	return "validates that a string contains valid scheme, host and optionally port and base path"
}

func (v UrlStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

//nolint:cyclop
func (UrlStringValidator) ValidateString(
	_ context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	diags := &resp.Diagnostics
	addErr := func(summary, detail string) {
		diags.AddAttributeError(req.Path, summary, detail)
	}

	u, err := url.Parse(req.ConfigValue.ValueString())
	if err != nil {
		addErr("Invalid URL", fmt.Sprintf("failed to parse URL: %s", err.Error()))

		return
	}

	if !slices.Contains([]string{"http", "https"}, u.Scheme) {
		detail := fmt.Sprintf(`URL must containt either "http" or "https" scheme, got: %q`, u.Scheme)

		addErr("Invalid URL Scheme", detail)
	}

	if u.Hostname() == "" {
		addErr("Missing URL Host", "URL must contain a hostname (either a domain name or an IP address)")
	}

	port := u.Port()
	if port != "" {
		if p, portErr := strconv.ParseInt(u.Port(), 10, 32); portErr != nil || p < 1 || p > 65535 {
			addErr("Invalid URL Port", "URL port must be an integer between 1 and 65535")
		}
	}

	if u.Path != "" {
		basePath := u.Path[1:]
		if len(basePath) > 255 || !urlBasePathRegexp.MatchString(strings.TrimLeft(basePath, "/")) {
			requirements := strings.Join(
				[]string{
					"- be at most 255 characters",
					"- not end with a slash (the / character)",
					"- consist only of alphanumeric and following characters: .-_[]+%*/",
				},
				"\n\t",
			)
			detail := fmt.Sprintf("URL path must adhere to the following requirements:\n\t%s", requirements)

			addErr("Invalid URL Path", detail)
		}
	}

	if u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		addErr("Invalid URL", "URL must not contain user information, query parameters or fragments")
	}
}

func urlFromParts(urlParts *UrlPartsModel) types.String {
	var urlString strings.Builder

	urlString.WriteString(urlParts.Scheme.ValueString())
	urlString.WriteString("://")
	urlString.WriteString(urlParts.Host.ValueString())

	if !urlParts.Port.IsNull() && !urlParts.Port.IsUnknown() {
		urlString.WriteRune(':')
		urlString.WriteString(strconv.Itoa(int(urlParts.Port.ValueInt32())))
	}

	if !urlParts.BasePath.IsNull() && !urlParts.BasePath.IsUnknown() {
		urlString.WriteRune('/')
		urlString.WriteString(urlParts.BasePath.ValueString())
	}

	return types.StringValue(urlString.String())
}
