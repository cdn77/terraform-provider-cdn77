package provider_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/oapi-codegen/nullable"
)

func TestAccCdnDataSource_NonExistingCdn(t *testing.T) {
	const cdnId = 7495732

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      acctest.Config(cdnDataSourceConfig, "id", cdnId),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`CDN Resource with id "%d" could not be found`, cdnId)),
			},
		},
	})
}

func TestAccCdnDataSource_OnlyRequiredFields(t *testing.T) {
	client := acctest.GetClient(t)

	originRequest := cdn77.OriginAddUrlJSONRequestBody{
		Host:   "my-totally-random-custom-host.com",
		Label:  "random origin",
		Scheme: "https",
	}
	originResponse, err := client.OriginAddUrlWithResponse(context.Background(), originRequest)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", originResponse, err)

	originId := originResponse.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, provider.OriginTypeUrl, originId)
	})

	const cdnLabel = "some cdn"

	cdnRequest := cdn77.CdnAddJSONRequestBody{Label: cdnLabel, OriginId: originId}
	cdnResponse, err := client.CdnAddWithResponse(context.Background(), cdnRequest)
	acctest.AssertResponseOk(t, "Failed to create CDN: %s", cdnResponse, err)

	cdnId := cdnResponse.JSON201.Id
	cdnCreationTime := cdnResponse.JSON201.CreationTime.Format(time.DateTime)
	cdnUrl := cdnResponse.JSON201.Url

	t.Cleanup(func() {
		acctest.MustDeleteCdn(t, client, cdnId)
	})

	attrEq := func(key, value string) resource.TestCheckFunc {
		return resource.TestCheckResourceAttr("data.cdn77_cdn.lorem", key, value)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: acctest.Config(cdnDataSourceConfig, "id", cdnId),
				Check: resource.ComposeAggregateTestCheckFunc(
					attrEq("id", fmt.Sprintf("%d", cdnId)),
					attrEq("cnames.#", "0"),
					attrEq("creation_time", cdnCreationTime),
					attrEq("label", cdnLabel),
					attrEq("origin_id", originId),
					attrEq("origin_protection_enabled", "false"),
					attrEq("url", cdnUrl),
					attrEq("cache.max_age", fmt.Sprintf("%d", cdn77.MaxAgeN17280)),
					attrEq("cache.requests_with_cookies_enabled", "true"),
					attrEq("secure_token.type", string(cdn77.SecureTokenTypeNone)),
					attrEq("query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeNone)),
					attrEq("headers.cors_enabled", "false"),
					attrEq("headers.cors_timing_enabled", "false"),
					attrEq("headers.cors_wildcard_enabled", "false"),
					attrEq("headers.host_header_forwarding_enabled", "false"),
					attrEq("headers.content_disposition_type", string(cdn77.ContentDispositionTypeNone)),
					attrEq("https_redirect.enabled", "false"),
					attrEq("mp4_pseudo_streaming_enabled", "false"),
					attrEq("quic_enabled", "false"),
					attrEq("waf_enabled", "false"),
					attrEq("ssl.type", string(cdn77.InstantSsl)),
					attrEq("hotlink_protection.type", string(cdn77.Disabled)),
					attrEq("hotlink_protection.empty_referer_denied", "false"),
					attrEq("ip_protection.type", string(cdn77.Disabled)),
					attrEq("geo_protection.type", string(cdn77.Disabled)),
					attrEq("rate_limit_enabled", "false"),
					attrEq("origin_headers.#", "0"),

					resource.TestCheckNoResourceAttr("data.cdn77_cdn.lorem", "note"),
					resource.TestCheckNoResourceAttr("data.cdn77_cdn.lorem", "cache.max_age_404"),
					resource.TestCheckNoResourceAttr("data.cdn77_cdn.lorem", "secure_token.token"),
					resource.TestCheckNoResourceAttr("data.cdn77_cdn.lorem", "query_string.parameters"),
					resource.TestCheckNoResourceAttr("data.cdn77_cdn.lorem", "https_redirect.code"),
					resource.TestCheckNoResourceAttr("data.cdn77_cdn.lorem", "ssl.ssl_id"),
					resource.TestCheckNoResourceAttr("data.cdn77_cdn.lorem", "stream"),
					resource.TestCheckNoResourceAttr("data.cdn77_cdn.lorem", "hotlink_protection.domains"),
					resource.TestCheckNoResourceAttr("data.cdn77_cdn.lorem", "ip_protection.ips"),
					resource.TestCheckNoResourceAttr("data.cdn77_cdn.lorem", "geo_protection.countries"),
				),
			},
		},
	})
}

func TestAccCdnDataSource_AllFields(t *testing.T) {
	client := acctest.GetClient(t)

	originRequest := cdn77.OriginAddUrlJSONRequestBody{
		Host:   "my-totally-random-custom-host.com",
		Label:  "random origin",
		Scheme: "https",
	}
	originResponse, err := client.OriginAddUrlWithResponse(context.Background(), originRequest)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", originResponse, err)

	originId := originResponse.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, provider.OriginTypeUrl, originId)
	})

	cdnCnames := []string{"my.cdn.com", "another-cname.example.com"}
	const cdnLabel = "some cdn"
	const cdnNote = "some note"

	sslId := acctest.MustAddSslWithCleanup(t, client, sslTestCert1, sslTestKey)

	cdnAddRequest := cdn77.CdnAddJSONRequestBody{
		Cnames:   util.Pointer(cdnCnames),
		Label:    cdnLabel,
		Note:     nullable.NewNullableWithValue(cdnNote),
		OriginId: originId,
	}
	cdnAddResponse, err := client.CdnAddWithResponse(context.Background(), cdnAddRequest)
	acctest.AssertResponseOk(t, "Failed to create CDN: %s", cdnAddResponse, err)

	cdnId := cdnAddResponse.JSON201.Id
	cdnCreationTime := cdnAddResponse.JSON201.CreationTime.Format(time.DateTime)
	cdnUrl := cdnAddResponse.JSON201.Url

	t.Cleanup(func() {
		acctest.MustDeleteCdn(t, client, cdnId)
	})

	cdnEditRequest := cdn77.CdnEditJSONRequestBody{
		Cache: &cdn77.Cache{
			MaxAge:                     util.Pointer(cdn77.MaxAgeN60),
			MaxAge404:                  nullable.NewNullableWithValue(cdn77.MaxAge404N5),
			RequestsWithCookiesEnabled: util.Pointer(false),
		},
		GeoProtection: &cdn77.GeoProtection{Countries: util.Pointer([]string{"CZ"}), Type: cdn77.Passlist},
		Headers: &cdn77.Headers{
			ContentDisposition: &cdn77.ContentDisposition{
				Type: util.Pointer(cdn77.ContentDispositionTypeParameter),
			},
			CorsEnabled:                 util.Pointer(true),
			CorsTimingEnabled:           util.Pointer(true),
			CorsWildcardEnabled:         util.Pointer(true),
			HostHeaderForwardingEnabled: util.Pointer(true),
		},
		HotlinkProtection: &cdn77.HotlinkProtection{
			Domains:            util.Pointer([]string{"xxx.cz"}),
			EmptyRefererDenied: true,
			Type:               cdn77.Passlist,
		},
		HttpsRedirect: &cdn77.HttpsRedirect{Code: util.Pointer(cdn77.N301), Enabled: true},
		IpProtection: &cdn77.IpProtection{
			Ips:  util.Pointer([]string{"1.1.1.1/32", "8.8.8.8/32"}),
			Type: cdn77.Blocklist,
		},
		OriginHeaders: &cdn77.OriginHeaders{
			Custom: nullable.NewNullableWithValue(map[string]string{"abc": "v1", "def": "v2"}),
		},
		QueryString: &cdn77.QueryString{
			IgnoreType: cdn77.QueryStringIgnoreTypeList,
			Parameters: util.Pointer([]string{"param"}),
		},
		Quic:        &cdn77.Quic{Enabled: true},
		RateLimit:   &cdn77.RateLimit{Enabled: true},
		SecureToken: &cdn77.SecureToken{Token: util.Pointer("abcd1234"), Type: cdn77.SecureTokenTypePath},
		Ssl:         &cdn77.CdnSsl{SslId: util.Pointer(sslId), Type: cdn77.SNI},
		Waf:         &cdn77.Waf{Enabled: true},
	}
	cdnEditResponse, err := client.CdnEditWithResponse(context.Background(), cdnId, cdnEditRequest)
	acctest.AssertResponseOk(t, "Failed to edit CDN: %s", cdnEditResponse, err)

	attrEq := func(key, value string) resource.TestCheckFunc {
		return resource.TestCheckResourceAttr("data.cdn77_cdn.lorem", key, value)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: acctest.Config(cdnDataSourceConfig, "id", cdnId),
				Check: resource.ComposeAggregateTestCheckFunc(
					attrEq("id", fmt.Sprintf("%d", cdnId)),
					attrEq("cnames.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.cdn77_cdn.lorem", "cnames.*", cdnCnames[0]),
					resource.TestCheckTypeSetElemAttr("data.cdn77_cdn.lorem", "cnames.*", cdnCnames[1]),
					attrEq("creation_time", cdnCreationTime),
					attrEq("label", cdnLabel),
					attrEq("note", cdnNote),
					attrEq("origin_id", originId),
					attrEq("origin_protection_enabled", "false"),
					attrEq("url", cdnUrl),
					attrEq("cache.max_age", fmt.Sprintf("%d", cdn77.MaxAgeN60)),
					attrEq("cache.max_age_404", fmt.Sprintf("%d", cdn77.MaxAge404N5)),
					attrEq("cache.requests_with_cookies_enabled", "false"),
					attrEq("secure_token.type", string(cdn77.SecureTokenTypePath)),
					attrEq("secure_token.token", "abcd1234"),
					attrEq("query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeList)),
					attrEq("query_string.parameters.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.cdn77_cdn.lorem", "query_string.parameters.*", "param"),
					attrEq("headers.cors_enabled", "true"),
					attrEq("headers.cors_timing_enabled", "true"),
					attrEq("headers.cors_wildcard_enabled", "true"),
					attrEq("headers.host_header_forwarding_enabled", "true"),
					attrEq("headers.content_disposition_type", string(cdn77.ContentDispositionTypeParameter)),
					attrEq("https_redirect.enabled", "true"),
					attrEq("https_redirect.code", fmt.Sprintf("%d", cdn77.N301)),
					attrEq("mp4_pseudo_streaming_enabled", "false"),
					attrEq("quic_enabled", "true"),
					attrEq("waf_enabled", "true"),
					attrEq("ssl.type", string(cdn77.SNI)),
					attrEq("ssl.ssl_id", sslId),
					attrEq("hotlink_protection.type", string(cdn77.Passlist)),
					attrEq("hotlink_protection.empty_referer_denied", "true"),
					attrEq("hotlink_protection.domains.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.cdn77_cdn.lorem", "hotlink_protection.domains.*", "xxx.cz"),
					attrEq("ip_protection.type", string(cdn77.Blocklist)),
					attrEq("ip_protection.ips.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.cdn77_cdn.lorem", "ip_protection.ips.*", "1.1.1.1/32"),
					resource.TestCheckTypeSetElemAttr("data.cdn77_cdn.lorem", "ip_protection.ips.*", "8.8.8.8/32"),
					attrEq("geo_protection.countries.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.cdn77_cdn.lorem", "geo_protection.countries.*", "CZ"),
					attrEq("geo_protection.type", string(cdn77.Passlist)),
					attrEq("rate_limit_enabled", "true"),
					attrEq("origin_headers.%", "2"),
					attrEq("origin_headers.abc", "v1"),
					attrEq("origin_headers.def", "v2"),

					resource.TestCheckNoResourceAttr("data.cdn77_cdn.lorem", "stream"),
				),
			},
			{
				PreConfig: func() {
					cdnEditRequest.Mp4PseudoStreaming = &cdn77.Mp4PseudoStreaming{Enabled: util.Pointer(true)}
					cdnEditRequest.QueryString.IgnoreType = cdn77.QueryStringIgnoreTypeAll
					cdnEditRequest.QueryString.Parameters = nil

					cdnEditResponse, err = client.CdnEditWithResponse(context.Background(), cdnId, cdnEditRequest)
					acctest.AssertResponseOk(t, "Failed to edit CDN: %s", cdnEditResponse, err)
				},
				Config: acctest.Config(cdnDataSourceConfig, "id", cdnId),
				Check: resource.ComposeAggregateTestCheckFunc(
					attrEq("query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeAll)),
					attrEq("query_string.#", "0"),
					attrEq("mp4_pseudo_streaming_enabled", "true"),
				),
			},
		},
	})
}

const cdnDataSourceConfig = `
data "cdn77_cdn" "lorem" {
  id = "{id}"
}
`
