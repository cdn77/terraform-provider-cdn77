package provider

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/mapping"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/origin"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &Cdn77Provider{}

type Cdn77Provider struct {
	// version is set to the provider version on release, "dev" when the provider is built and ran locally,
	// and "test" when running acceptance testing.
	version string
}

type Cdn77ProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
	Timeout  types.Int64  `tfsdk:"timeout"`
}

func (p *Cdn77Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cdn77"
	resp.Version = p.version
}

func (*Cdn77Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "API endpoint; defaults to https://api.cdn77.com",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Authentication token from https://client.cdn77.com/account/api",
				Optional:            true,
				Sensitive:           true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for all API calls (in seconds). Negative values disable the timeout. " +
					"Default is 30 seconds.",
				Optional: true,
			},
		},
		Description: "The CDN77 provider is used to interact with CDN77 resources.",
	}
}

func (p *Cdn77Provider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var data Cdn77ProviderModel
	if resp.Diagnostics.Append(req.Config.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}

	if data.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown CDN77 API endpoint",
			"The provider cannot create the CDN77 API client as there is an unknown configuration value for the "+
				"API endpoint. Either target apply the source of the value first, set the value statically in the "+
				"configuration, or use the CDN77_ENDPOINT environment variable.",
		)
	}

	if data.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown CDN77 API token",
			"The provider cannot create the CDN77 API client as there is an unknown configuration value for the "+
				"API token. Either target apply the source of the value first, set the value statically in the "+
				"configuration, or use the CDN77_TOKEN environment variable.",
		)
	}

	if data.Timeout.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown CDN77 API timeout",
			"The provider cannot create the CDN77 API client as there is an unknown configuration value for the "+
				"API timeout. Either target apply the source of the value first, set the value statically in the "+
				"configuration, or use the CDN77_TIMEOUT environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override with Terraform configuration value if set.
	endpoint, token, timeout := p.getConfig(&resp.Diagnostics, data)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := NewClient(endpoint, token, timeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to initialize CDN77 API client", err.Error())

		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (*Cdn77Provider) getConfig(
	diags *diag.Diagnostics,
	data Cdn77ProviderModel,
) (endpoint string, token string, timeout time.Duration) {
	endpoint = os.Getenv("CDN77_ENDPOINT")
	token = os.Getenv("CDN77_TOKEN")

	timeoutString := os.Getenv("CDN77_TIMEOUT")
	if timeoutString != "" {
		timeoutSeconds, err := strconv.Atoi(timeoutString)
		if err != nil {
			diags.AddError(
				"Invalid CDN77 API timeout",
				fmt.Sprintf("Failed to convert environment variable CDN77_TIMEOUT value to an integer: %s", err),
			)
		}

		timeout = time.Duration(timeoutSeconds)
	}

	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}

	if endpoint == "" {
		endpoint = "https://api.cdn77.com"
	}

	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	if token == "" {
		diags.AddAttributeError(
			path.Root("token"),
			"Missing CDN77 API token",
			"The provider cannot create the CDN77 API client because API token is not set. "+
				"Either set the value statically in the configuration, or use the CDN77_TOKEN environment variable.",
		)
	}

	if !data.Timeout.IsNull() {
		timeout = time.Duration(data.Timeout.ValueInt64())
	}

	if timeout == 0 {
		timeout = 30
	}

	timeout *= time.Second

	return endpoint, token, timeout
}

func (*Cdn77Provider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		mapping.ResourceFactory(mapping.Cdn),
		mapping.ResourceFactory(mapping.OriginAws),
		mapping.ResourceFactory(mapping.OriginObjectStorage),
		mapping.ResourceFactory(mapping.OriginUrl),
		mapping.ResourceFactory(mapping.Ssl),
	}
}

func (*Cdn77Provider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		mapping.DataSourceFactory(mapping.Cdn),
		mapping.DataSourceFactory(mapping.Cdns),
		mapping.DataSourceFactory(mapping.ObjectStorages),
		mapping.DataSourceFactory(mapping.OriginAws),
		mapping.DataSourceFactory(mapping.OriginObjectStorage),
		mapping.DataSourceFactory(mapping.OriginUrl),
		origin.NewAllDataSource,
		mapping.DataSourceFactory(mapping.Ssl),
		mapping.DataSourceFactory(mapping.Ssls),
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Cdn77Provider{
			version: version,
		}
	}
}

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (fn RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func NewClient(endpoint string, token string, timeout time.Duration) (cdn77.ClientWithResponsesInterface, error) {
	dialer := &net.Dialer{}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          10,
		MaxConnsPerHost:       10,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client, err := cdn77.NewClientWithResponses(
		endpoint,
		cdn77.WithHTTPClient(&http.Client{Transport: transport, Timeout: timeout}),
		cdn77.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

type StateProvider interface {
	Get(ctx context.Context, target any) diag.Diagnostics
}
