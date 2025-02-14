package cdn_test

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest/testdata"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/origin"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oapi-codegen/nullable"
)

func TestAccCdnResource(t *testing.T) {
	const rsc = "cdn77_cdn.lorem"
	const originRsc = "cdn77_origin_url.url"

	client := acctest.GetClient(t)
	var cdnId string
	var originId string
	var cdnCreationTime string
	var cdnUrl string
	sslId := acctest.MustAddSslWithCleanup(t, client, testdata.SslCert1, testdata.SslKey)

	attrEq := func(key, value string) resource.TestCheckFunc {
		return resource.TestCheckResourceAttr(rsc, key, value)
	}

	acctest.Run(t, checkCdnsAndOriginDestroyed(client),
		resource.TestStep{
			Config: OriginResourceConfig + `resource "cdn77_cdn" "lorem" {
				label = "my cdn"
				origin_id = cdn77_origin_url.url.id
			}`,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(originRsc, plancheck.ResourceActionCreate),
					plancheck.ExpectResourceAction(rsc, plancheck.ResourceActionCreate),
				},
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAndAssignAttr(originRsc, "id", &originId),
				acctest.CheckAndAssignAttr(rsc, "id", &cdnId),
				attrEq("label", "my cdn"),
				acctest.CheckAttr(rsc, "origin_id", &originId),
				acctest.CheckAndAssignAttr(rsc, "creation_time", &cdnCreationTime),
				acctest.CheckAndAssignAttr(rsc, "url", &cdnUrl),
				attrEq("cache.max_age", fmt.Sprintf("%d", cdn77.MaxAgeN17280)),
				attrEq("cache.requests_with_cookies_enabled", "true"),
				attrEq("cnames.#", "0"),
				attrEq("geo_protection.type", string(cdn77.Disabled)),
				attrEq("headers.content_disposition_type", string(cdn77.ContentDispositionTypeNone)),
				attrEq("headers.cors_enabled", "false"),
				attrEq("headers.cors_timing_enabled", "false"),
				attrEq("headers.cors_wildcard_enabled", "false"),
				attrEq("headers.host_header_forwarding_enabled", "false"),
				attrEq("hotlink_protection.empty_referer_denied", "false"),
				attrEq("hotlink_protection.type", string(cdn77.Disabled)),
				attrEq("https_redirect.enabled", "false"),
				attrEq("ip_protection.type", string(cdn77.Disabled)),
				attrEq("mp4_pseudo_streaming_enabled", "false"),
				attrEq("origin_headers.#", "0"),
				attrEq("query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeNone)),
				attrEq("rate_limit_enabled", "false"),
				attrEq("secure_token.type", string(cdn77.SecureTokenTypeNone)),
				attrEq("ssl.type", string(cdn77.InstantSsl)),

				resource.TestCheckNoResourceAttr(rsc, "stream"),
				resource.TestCheckNoResourceAttr(rsc, "cache.max_age_404"),
				resource.TestCheckNoResourceAttr(rsc, "geo_protection.countries"),
				resource.TestCheckNoResourceAttr(rsc, "hotlink_protection.domains"),
				resource.TestCheckNoResourceAttr(rsc, "https_redirect.code"),
				resource.TestCheckNoResourceAttr(rsc, "ip_protection.ips"),
				resource.TestCheckNoResourceAttr(rsc, "note"),
				resource.TestCheckNoResourceAttr(rsc, "query_string.parameters"),
				resource.TestCheckNoResourceAttr(rsc, "secure_token.token"),
				resource.TestCheckNoResourceAttr(rsc, "ssl.ssl_id"),

				checkCdnDefaults(client, &cdnId, &originId, "my cdn"),
			),
		},
		resource.TestStep{
			Config: OriginResourceConfig + `resource "cdn77_cdn" "lorem" {
				label = "changed the label"
				origin_id = cdn77_origin_url.url.id
				cache = {
					max_age = 60
					max_age_404 = 5
					requests_with_cookies_enabled = false
				}
				cnames = ["my.cdn.cz", "other.cdn.com"]
				headers = {
					cors_enabled = true
					cors_timing_enabled = true
					cors_wildcard_enabled = true
					host_header_forwarding_enabled = true
					content_disposition_type = "parameter"
				}
				https_redirect = {
					enabled = true
					code = 301
				}
				note = "custom note"
				query_string = {
					ignore_type = "list"
					parameters = ["param"]
				}
				secure_token = {
					type = "path"
					token = "abcd1234"
				}
			}`,
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionUpdate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAttr(rsc, "id", &cdnId),
				attrEq("label", "changed the label"),
				acctest.CheckAttr(rsc, "origin_id", &originId),
				acctest.CheckAttr(rsc, "creation_time", &cdnCreationTime),
				acctest.CheckAttr(rsc, "url", &cdnUrl),
				attrEq("cache.max_age", fmt.Sprintf("%d", cdn77.MaxAgeN60)),
				attrEq("cache.max_age_404", fmt.Sprintf("%d", cdn77.MaxAge404N5)),
				attrEq("cache.requests_with_cookies_enabled", "false"),
				attrEq("cnames.#", "2"),
				resource.TestCheckTypeSetElemAttr(rsc, "cnames.*", "my.cdn.cz"),
				resource.TestCheckTypeSetElemAttr(rsc, "cnames.*", "other.cdn.com"),
				attrEq("geo_protection.type", string(cdn77.Disabled)),
				attrEq("headers.content_disposition_type", string(cdn77.ContentDispositionTypeParameter)),
				attrEq("headers.cors_enabled", "true"),
				attrEq("headers.cors_timing_enabled", "true"),
				attrEq("headers.cors_wildcard_enabled", "true"),
				attrEq("headers.host_header_forwarding_enabled", "true"),
				attrEq("hotlink_protection.empty_referer_denied", "false"),
				attrEq("hotlink_protection.type", string(cdn77.Disabled)),
				attrEq("https_redirect.code", fmt.Sprintf("%d", cdn77.N301)),
				attrEq("https_redirect.enabled", "true"),
				attrEq("ip_protection.type", string(cdn77.Disabled)),
				attrEq("mp4_pseudo_streaming_enabled", "false"),
				attrEq("note", "custom note"),
				attrEq("origin_headers.#", "0"),
				attrEq("query_string.parameters.#", "1"),
				resource.TestCheckTypeSetElemAttr(rsc, "query_string.parameters.*", "param"),
				attrEq("query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeList)),
				attrEq("rate_limit_enabled", "false"),
				attrEq("secure_token.token", "abcd1234"),
				attrEq("secure_token.type", string(cdn77.SecureTokenTypePath)),
				attrEq("ssl.type", string(cdn77.InstantSsl)),

				resource.TestCheckNoResourceAttr(rsc, "stream"),
				resource.TestCheckNoResourceAttr(rsc, "geo_protection.countries"),
				resource.TestCheckNoResourceAttr(rsc, "hotlink_protection.domains"),
				resource.TestCheckNoResourceAttr(rsc, "ip_protection.ips"),
				resource.TestCheckNoResourceAttr(rsc, "ssl.ssl_id"),

				checkCdn(client, &cdnId, func(c *cdn77.Cdn) error {
					slices.SortStableFunc(c.Cnames, func(a, b cdn77.Cname) int {
						return cmp.Compare(a.Cname, b.Cname)
					})

					return errors.Join(
						acctest.EqualField("label", c.Label, "changed the label"),
						acctest.NullFieldEqual("origin_id", c.OriginId, originId),
						acctest.EqualField("cache.max_age", *c.Cache.MaxAge, cdn77.MaxAgeN60),
						acctest.NullFieldEqual("cache.max_age_404", c.Cache.MaxAge404, cdn77.MaxAge404N5),
						acctest.EqualField(
							"cache.requests_with_cookies_enabled",
							*c.Cache.RequestsWithCookiesEnabled,
							false,
						),
						acctest.EqualField("cnames.#", len(c.Cnames), 2),
						acctest.EqualField("cnames.0", c.Cnames[0].Cname, "my.cdn.cz"),
						acctest.EqualField("cnames.1", c.Cnames[1].Cname, "other.cdn.com"),
						acctest.EqualField("headers.cors_enabled", *c.Headers.CorsEnabled, true),
						acctest.EqualField("headers.cors_timing_enabled", *c.Headers.CorsTimingEnabled, true),
						acctest.EqualField("headers.cors_wildcard_enabled", *c.Headers.CorsWildcardEnabled, true),
						acctest.EqualField(
							"headers.host_header_forwarding_enabled",
							*c.Headers.HostHeaderForwardingEnabled,
							true,
						),
						acctest.EqualField(
							"headers.content_disposition_type",
							*c.Headers.ContentDisposition.Type,
							cdn77.ContentDispositionTypeParameter,
						),
						acctest.EqualField("https_redirect.enabled", c.HttpsRedirect.Enabled, true),
						acctest.EqualField("https_redirect.code", *c.HttpsRedirect.Code, cdn77.N301),
						acctest.NullFieldEqual("note", c.Note, "custom note"),
						acctest.EqualField(
							"query_string.ignore_type",
							c.QueryString.IgnoreType,
							cdn77.QueryStringIgnoreTypeList,
						),
						acctest.EqualField("query_string.parameters.#", len(*c.QueryString.Parameters), 1),
						acctest.EqualField("query_string.parameters.0", (*c.QueryString.Parameters)[0], "param"),
						acctest.EqualField("secure_token.type", c.SecureToken.Type, cdn77.SecureTokenTypePath),
						acctest.EqualField("secure_token.token", *c.SecureToken.Token, "abcd1234"),
					)
				}),
			),
		},
		resource.TestStep{
			Config: OriginResourceConfig + acctest.Config(
				`resource "cdn77_cdn" "lorem" {
					label = "changed the label"
					origin_id = cdn77_origin_url.url.id
					cache = {
						max_age = 60
						max_age_404 = 5
						requests_with_cookies_enabled = false
					}
					cnames = ["my.cdn.cz", "other.cdn.com"]
					geo_protection = {
						type = "blocklist"
						countries = ["SK"]
					}
					headers = {
						cors_enabled = true
						cors_timing_enabled = true
						cors_wildcard_enabled = true
						host_header_forwarding_enabled = true
						content_disposition_type = "parameter"
					}
					hotlink_protection = {
						type = "blocklist"
						domains = ["example.com"]
						empty_referer_denied = true
					}
					https_redirect = {
						enabled = true
						code = 301
					}
					ip_protection = {
						type = "passlist"
						ips = ["1.1.1.1/32", "8.8.8.8/32"]
					}
					mp4_pseudo_streaming_enabled = true
					note = "custom note"
					origin_headers = {
						abc = "v1"
						def = "v2"
					}
					query_string = {
						ignore_type = "all"
					}
					rate_limit_enabled = true
					secure_token = {
						type = "path"
						token = "abcd1234"
					}
					ssl = {
						type = "SNI"
						ssl_id = "{sslId}"
					}
				}`,
				"sslId", sslId,
			),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionUpdate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAttr(rsc, "id", &cdnId),
				attrEq("label", "changed the label"),
				acctest.CheckAttr(rsc, "origin_id", &originId),
				acctest.CheckAttr(rsc, "creation_time", &cdnCreationTime),
				acctest.CheckAttr(rsc, "url", &cdnUrl),
				resource.TestCheckResourceAttrWith(rsc, "creation_time", func(value string) (err error) {
					return acctest.Equal(value, cdnCreationTime)
				}),
				attrEq("cache.max_age", fmt.Sprintf("%d", cdn77.MaxAgeN60)),
				attrEq("cache.max_age_404", fmt.Sprintf("%d", cdn77.MaxAge404N5)),
				attrEq("cache.requests_with_cookies_enabled", "false"),
				attrEq("cnames.#", "2"),
				resource.TestCheckTypeSetElemAttr(rsc, "cnames.*", "my.cdn.cz"),
				resource.TestCheckTypeSetElemAttr(rsc, "cnames.*", "other.cdn.com"),
				attrEq("geo_protection.countries.#", "1"),
				resource.TestCheckTypeSetElemAttr(rsc, "geo_protection.countries.*", "SK"),
				attrEq("geo_protection.type", string(cdn77.Blocklist)),
				attrEq("headers.content_disposition_type", string(cdn77.ContentDispositionTypeParameter)),
				attrEq("headers.cors_enabled", "true"),
				attrEq("headers.cors_timing_enabled", "true"),
				attrEq("headers.cors_wildcard_enabled", "true"),
				attrEq("headers.host_header_forwarding_enabled", "true"),
				attrEq("hotlink_protection.domains.#", "1"),
				resource.TestCheckTypeSetElemAttr(rsc, "hotlink_protection.domains.*", "example.com"),
				attrEq("hotlink_protection.empty_referer_denied", "true"),
				attrEq("hotlink_protection.type", string(cdn77.Blocklist)),
				attrEq("https_redirect.code", fmt.Sprintf("%d", cdn77.N301)),
				attrEq("https_redirect.enabled", "true"),
				attrEq("ip_protection.ips.#", "2"),
				resource.TestCheckTypeSetElemAttr(rsc, "ip_protection.ips.*", "1.1.1.1/32"),
				resource.TestCheckTypeSetElemAttr(rsc, "ip_protection.ips.*", "8.8.8.8/32"),
				attrEq("ip_protection.type", string(cdn77.Passlist)),
				attrEq("mp4_pseudo_streaming_enabled", "true"),
				attrEq("note", "custom note"),
				attrEq("origin_headers.%", "2"),
				attrEq("origin_headers.abc", "v1"),
				attrEq("origin_headers.def", "v2"),
				attrEq("query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeAll)),
				attrEq("query_string.parameters.#", "0"),
				attrEq("rate_limit_enabled", "true"),
				attrEq("secure_token.token", "abcd1234"),
				attrEq("secure_token.type", string(cdn77.SecureTokenTypePath)),
				attrEq("ssl.ssl_id", sslId),
				attrEq("ssl.type", string(cdn77.SNI)),

				resource.TestCheckNoResourceAttr(rsc, "stream"),

				checkCdn(client, &cdnId, func(c *cdn77.Cdn) error {
					slices.SortStableFunc(c.Cnames, func(a, b cdn77.Cname) int {
						return cmp.Compare(a.Cname, b.Cname)
					})
					slices.SortStableFunc(*c.IpProtection.Ips, cmp.Compare)

					return errors.Join(
						acctest.EqualField("label", c.Label, "changed the label"),
						acctest.NullFieldEqual("origin_id", c.OriginId, originId),
						acctest.EqualField("cache.max_age", *c.Cache.MaxAge, cdn77.MaxAgeN60),
						acctest.NullFieldEqual("cache.max_age_404", c.Cache.MaxAge404, cdn77.MaxAge404N5),
						acctest.EqualField(
							"cache.requests_with_cookies_enabled",
							*c.Cache.RequestsWithCookiesEnabled,
							false,
						),
						acctest.EqualField("cnames.#", len(c.Cnames), 2),
						acctest.EqualField("cnames.0", c.Cnames[0].Cname, "my.cdn.cz"),
						acctest.EqualField("cnames.1", c.Cnames[1].Cname, "other.cdn.com"),
						acctest.EqualField("geo_protection.type", c.GeoProtection.Type, cdn77.Blocklist),
						acctest.EqualField("geo_protection.countries.0", (*c.GeoProtection.Countries)[0], "SK"),
						acctest.EqualField("headers.cors_enabled", *c.Headers.CorsEnabled, true),
						acctest.EqualField("headers.cors_timing_enabled", *c.Headers.CorsTimingEnabled, true),
						acctest.EqualField("headers.cors_wildcard_enabled", *c.Headers.CorsWildcardEnabled, true),
						acctest.EqualField(
							"headers.host_header_forwarding_enabled",
							*c.Headers.HostHeaderForwardingEnabled,
							true,
						),
						acctest.EqualField(
							"headers.content_disposition_type",
							*c.Headers.ContentDisposition.Type,
							cdn77.ContentDispositionTypeParameter,
						),
						acctest.EqualField("hotlink_protection.type", c.HotlinkProtection.Type, cdn77.Blocklist),
						acctest.EqualField("hotlink_protection.domains.#", len(*c.HotlinkProtection.Domains), 1),
						acctest.EqualField(
							"hotlink_protection.domains.0",
							(*c.HotlinkProtection.Domains)[0],
							"example.com",
						),
						acctest.EqualField(
							"hotlink_protection.empty_referer_denied",
							c.HotlinkProtection.EmptyRefererDenied,
							true,
						),
						acctest.EqualField("https_redirect.enabled", c.HttpsRedirect.Enabled, true),
						acctest.EqualField("https_redirect.code", *c.HttpsRedirect.Code, cdn77.N301),
						acctest.EqualField("ip_protection.type", c.IpProtection.Type, cdn77.Passlist),
						acctest.EqualField("ip_protection.ips.#", len(*c.IpProtection.Ips), 2),
						acctest.EqualField("ip_protection.ips.0", (*c.IpProtection.Ips)[0], "1.1.1.1/32"),
						acctest.EqualField("ip_protection.ips.1", (*c.IpProtection.Ips)[1], "8.8.8.8/32"),
						acctest.EqualField("mp4_pseudo_streaming_enabled", *c.Mp4PseudoStreaming.Enabled, true),
						acctest.NullFieldEqual("note", c.Note, "custom note"),
						func() error {
							expected := map[string]string{"abc": "v1", "def": "v2"}

							if value, err := c.OriginHeaders.Custom.Get(); err == nil {
								if len(value) == len(expected) {
									valid := true

									for k, v := range expected {
										if value[k] != v {
											valid = false
										}
									}

									if valid {
										return nil
									}
								}
							}

							return fmt.Errorf(
								"field origin_headers: expected %+v, got: %+v",
								nullable.NewNullableWithValue(expected),
								c.OriginHeaders.Custom,
							)
						}(),
						acctest.EqualField(
							"query_string.ignore_type",
							c.QueryString.IgnoreType,
							cdn77.QueryStringIgnoreTypeAll,
						),
						acctest.EqualField("query_string.parameters", c.QueryString.Parameters, nil),
						acctest.EqualField("rate_limit_enabled", c.RateLimit.Enabled, true),
						acctest.EqualField("secure_token.type", c.SecureToken.Type, cdn77.SecureTokenTypePath),
						acctest.EqualField("secure_token.token", *c.SecureToken.Token, "abcd1234"),
						acctest.EqualField("ssl.type", c.Ssl.Type, cdn77.SNI),
						acctest.EqualField("ssl.ssl_id", *c.Ssl.SslId, sslId),
					)
				}),
			),
		},
		resource.TestStep{
			Config: OriginResourceConfig + `resource "cdn77_cdn" "lorem" {
				label = "my cdn"
				origin_id = cdn77_origin_url.url.id
			}`,
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionUpdate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAttr(originRsc, "id", &originId),
				acctest.CheckAttr(rsc, "id", &cdnId),
				attrEq("label", "my cdn"),
				acctest.CheckAttr(rsc, "origin_id", &originId),
				acctest.CheckAttr(rsc, "creation_time", &cdnCreationTime),
				acctest.CheckAttr(rsc, "url", &cdnUrl),
				attrEq("cache.max_age", fmt.Sprintf("%d", cdn77.MaxAgeN17280)),
				attrEq("cache.requests_with_cookies_enabled", "true"),
				attrEq("cnames.#", "0"),
				attrEq("geo_protection.type", string(cdn77.Disabled)),
				attrEq("headers.content_disposition_type", string(cdn77.ContentDispositionTypeNone)),
				attrEq("headers.cors_enabled", "false"),
				attrEq("headers.cors_timing_enabled", "false"),
				attrEq("headers.cors_wildcard_enabled", "false"),
				attrEq("headers.host_header_forwarding_enabled", "false"),
				attrEq("hotlink_protection.empty_referer_denied", "false"),
				attrEq("hotlink_protection.type", string(cdn77.Disabled)),
				attrEq("https_redirect.enabled", "false"),
				attrEq("ip_protection.type", string(cdn77.Disabled)),
				attrEq("mp4_pseudo_streaming_enabled", "false"),
				attrEq("origin_headers.#", "0"),
				attrEq("query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeNone)),
				attrEq("rate_limit_enabled", "false"),
				attrEq("secure_token.type", string(cdn77.SecureTokenTypeNone)),
				attrEq("ssl.type", string(cdn77.InstantSsl)),

				resource.TestCheckNoResourceAttr(rsc, "stream"),
				resource.TestCheckNoResourceAttr(rsc, "cache.max_age_404"),
				resource.TestCheckNoResourceAttr(rsc, "geo_protection.countries"),
				resource.TestCheckNoResourceAttr(rsc, "hotlink_protection.domains"),
				resource.TestCheckNoResourceAttr(rsc, "https_redirect.code"),
				resource.TestCheckNoResourceAttr(rsc, "ip_protection.ips"),
				resource.TestCheckNoResourceAttr(rsc, "note"),
				resource.TestCheckNoResourceAttr(rsc, "query_string.parameters"),
				resource.TestCheckNoResourceAttr(rsc, "secure_token.token"),
				resource.TestCheckNoResourceAttr(rsc, "ssl.ssl_id"),

				checkCdnDefaults(client, &cdnId, &originId, "my cdn"),
			),
		},
	)
}

func TestAccCdnResource_Import(t *testing.T) {
	client := acctest.GetClient(t)
	rsc := "cdn77_cdn.lorem"
	var cdnId string

	acctest.Run(t, checkCdnsAndOriginDestroyed(client),
		resource.TestStep{
			Config: OriginResourceConfig + `resource "cdn77_cdn" "lorem" {
					origin_id = cdn77_origin_url.url.id
					label = "my cdn"
					note = "custom note"
				}`,
			Check: acctest.CheckAndAssignAttr(rsc, "id", &cdnId),
		},
		resource.TestStep{
			ResourceName: rsc,
			ImportState:  true,
			ImportStateIdFunc: func(*terraform.State) (string, error) {
				return cdnId, nil
			},
			ImportStateVerify: true,
		},
	)
}

func TestAccCdnDataSource_OnlyRequiredFields(t *testing.T) {
	const rsc = "data.cdn77_cdn.lorem"
	const nonExistingCdnId = 7495732
	client := acctest.GetClient(t)

	originRequest := cdn77.OriginCreateUrlJSONRequestBody{
		Label:  "random origin",
		Scheme: "https",
		Host:   "my-totally-random-custom-host.com",
	}
	originResponse, err := client.OriginCreateUrlWithResponse(t.Context(), originRequest)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", originResponse, err)

	originId := originResponse.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, origin.TypeUrl, originId)
	})

	const cdnLabel = "some cdn"

	cdnRequest := cdn77.CdnAddJSONRequestBody{Label: cdnLabel, OriginId: originId}
	cdnResponse, err := client.CdnAddWithResponse(t.Context(), cdnRequest)
	acctest.AssertResponseOk(t, "Failed to create CDN: %s", cdnResponse, err)

	cdnId := cdnResponse.JSON201.Id
	cdnCreationTime := cdnResponse.JSON201.CreationTime.Format(time.DateTime)
	cdnUrl := cdnResponse.JSON201.Url

	t.Cleanup(func() {
		acctest.MustDeleteCdn(t, client, cdnId)
	})

	attrEq := func(key, value string) resource.TestCheckFunc {
		return resource.TestCheckResourceAttr(rsc, key, value)
	}

	acctest.Run(t, nil,
		resource.TestStep{
			Config: acctest.Config(cdnDataSourceConfig, "id", nonExistingCdnId),
			ExpectError: regexp.MustCompile(
				fmt.Sprintf(`CDN Resource with id "%d" could not be found`, nonExistingCdnId),
			),
		},
		resource.TestStep{
			Config: acctest.Config(cdnDataSourceConfig, "id", cdnId),
			Check: resource.ComposeAggregateTestCheckFunc(
				attrEq("id", fmt.Sprintf("%d", cdnId)),
				attrEq("label", cdnLabel),
				attrEq("origin_id", originId),
				attrEq("creation_time", cdnCreationTime),
				attrEq("url", cdnUrl),
				attrEq("cache.max_age", fmt.Sprintf("%d", cdn77.MaxAgeN17280)),
				attrEq("cache.requests_with_cookies_enabled", "true"),
				attrEq("cnames.#", "0"),
				attrEq("geo_protection.type", string(cdn77.Disabled)),
				attrEq("headers.cors_enabled", "false"),
				attrEq("headers.cors_timing_enabled", "false"),
				attrEq("headers.cors_wildcard_enabled", "false"),
				attrEq("headers.host_header_forwarding_enabled", "false"),
				attrEq("headers.content_disposition_type", string(cdn77.ContentDispositionTypeNone)),
				attrEq("hotlink_protection.type", string(cdn77.Disabled)),
				attrEq("hotlink_protection.empty_referer_denied", "false"),
				attrEq("https_redirect.enabled", "false"),
				attrEq("ip_protection.type", string(cdn77.Disabled)),
				attrEq("mp4_pseudo_streaming_enabled", "false"),
				attrEq("origin_headers.#", "0"),
				attrEq("query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeNone)),
				attrEq("rate_limit_enabled", "false"),
				attrEq("secure_token.type", string(cdn77.SecureTokenTypeNone)),
				attrEq("ssl.type", string(cdn77.InstantSsl)),

				resource.TestCheckNoResourceAttr(rsc, "stream"),
				resource.TestCheckNoResourceAttr(rsc, "cache.max_age_404"),
				resource.TestCheckNoResourceAttr(rsc, "geo_protection.countries"),
				resource.TestCheckNoResourceAttr(rsc, "hotlink_protection.domains"),
				resource.TestCheckNoResourceAttr(rsc, "https_redirect.code"),
				resource.TestCheckNoResourceAttr(rsc, "ip_protection.ips"),
				resource.TestCheckNoResourceAttr(rsc, "note"),
				resource.TestCheckNoResourceAttr(rsc, "query_string.parameters"),
				resource.TestCheckNoResourceAttr(rsc, "secure_token.token"),
				resource.TestCheckNoResourceAttr(rsc, "ssl.ssl_id"),
			),
		},
	)
}

func TestAccCdnDataSource_AllFields(t *testing.T) {
	const rsc = "data.cdn77_cdn.lorem"
	client := acctest.GetClient(t)

	originRequest := cdn77.OriginCreateUrlJSONRequestBody{
		Label:  "random origin",
		Scheme: "https",
		Host:   "my-totally-random-custom-host.com",
	}
	originResponse, err := client.OriginCreateUrlWithResponse(t.Context(), originRequest)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", originResponse, err)

	originId := originResponse.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, origin.TypeUrl, originId)
	})

	cdnCnames := []string{"my.cdn.com", "another-cname.example.com"}
	const cdnLabel = "some cdn"
	const cdnNote = "some note"

	sslId := acctest.MustAddSslWithCleanup(t, client, testdata.SslCert1, testdata.SslKey)

	cdnAddRequest := cdn77.CdnAddJSONRequestBody{
		Label:    cdnLabel,
		OriginId: originId,
		Cnames:   util.Pointer(cdnCnames),
		Note:     nullable.NewNullableWithValue(cdnNote),
	}
	cdnAddResponse, err := client.CdnAddWithResponse(t.Context(), cdnAddRequest)
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
		RateLimit:   &cdn77.RateLimit{Enabled: true},
		SecureToken: &cdn77.SecureToken{Token: util.Pointer("abcd1234"), Type: cdn77.SecureTokenTypePath},
		Ssl:         &cdn77.CdnSsl{SslId: util.Pointer(sslId), Type: cdn77.SNI},
	}
	cdnEditResponse, err := client.CdnEditWithResponse(t.Context(), cdnId, cdnEditRequest)
	acctest.AssertResponseOk(t, "Failed to edit CDN: %s", cdnEditResponse, err)

	attrEq := func(key, value string) resource.TestCheckFunc {
		return resource.TestCheckResourceAttr(rsc, key, value)
	}

	acctest.Run(t, nil, resource.TestStep{
		Config: acctest.Config(cdnDataSourceConfig, "id", cdnId),
		Check: resource.ComposeAggregateTestCheckFunc(
			attrEq("id", fmt.Sprintf("%d", cdnId)),
			attrEq("label", cdnLabel),
			attrEq("origin_id", originId),
			attrEq("creation_time", cdnCreationTime),
			attrEq("url", cdnUrl),
			attrEq("cache.max_age", fmt.Sprintf("%d", cdn77.MaxAgeN60)),
			attrEq("cache.max_age_404", fmt.Sprintf("%d", cdn77.MaxAge404N5)),
			attrEq("cache.requests_with_cookies_enabled", "false"),
			attrEq("cnames.#", "2"),
			resource.TestCheckTypeSetElemAttr(rsc, "cnames.*", cdnCnames[0]),
			resource.TestCheckTypeSetElemAttr(rsc, "cnames.*", cdnCnames[1]),
			attrEq("geo_protection.countries.#", "1"),
			resource.TestCheckTypeSetElemAttr(rsc, "geo_protection.countries.*", "CZ"),
			attrEq("geo_protection.type", string(cdn77.Passlist)),
			attrEq("headers.content_disposition_type", string(cdn77.ContentDispositionTypeParameter)),
			attrEq("headers.cors_enabled", "true"),
			attrEq("headers.cors_timing_enabled", "true"),
			attrEq("headers.cors_wildcard_enabled", "true"),
			attrEq("headers.host_header_forwarding_enabled", "true"),
			attrEq("hotlink_protection.domains.#", "1"),
			resource.TestCheckTypeSetElemAttr(rsc, "hotlink_protection.domains.*", "xxx.cz"),
			attrEq("hotlink_protection.empty_referer_denied", "true"),
			attrEq("hotlink_protection.type", string(cdn77.Passlist)),
			attrEq("https_redirect.code", fmt.Sprintf("%d", cdn77.N301)),
			attrEq("https_redirect.enabled", "true"),
			attrEq("ip_protection.ips.#", "2"),
			resource.TestCheckTypeSetElemAttr(rsc, "ip_protection.ips.*", "1.1.1.1/32"),
			resource.TestCheckTypeSetElemAttr(rsc, "ip_protection.ips.*", "8.8.8.8/32"),
			attrEq("ip_protection.type", string(cdn77.Blocklist)),
			attrEq("mp4_pseudo_streaming_enabled", "false"),
			attrEq("note", cdnNote),
			attrEq("origin_headers.%", "2"),
			attrEq("origin_headers.abc", "v1"),
			attrEq("origin_headers.def", "v2"),
			attrEq("query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeList)),
			attrEq("query_string.parameters.#", "1"),
			resource.TestCheckTypeSetElemAttr(rsc, "query_string.parameters.*", "param"),
			attrEq("rate_limit_enabled", "true"),
			attrEq("secure_token.token", "abcd1234"),
			attrEq("secure_token.type", string(cdn77.SecureTokenTypePath)),
			attrEq("ssl.ssl_id", sslId),
			attrEq("ssl.type", string(cdn77.SNI)),

			resource.TestCheckNoResourceAttr(rsc, "stream"),
		),
	})
}

func checkCdn(
	client cdn77.ClientWithResponsesInterface,
	cdnId *string,
	fn func(o *cdn77.Cdn) error,
) resource.TestCheckFunc {
	return func(*terraform.State) error {
		cdnIdInt, err := strconv.Atoi(*cdnId)
		if err != nil {
			return fmt.Errorf("failed to convert CDN ID to int: %w", err)
		}

		response, err := client.CdnDetailWithResponse(context.Background(), cdnIdInt)
		message := fmt.Sprintf("failed to get CDN[id=%d]: %%s", cdnIdInt)

		if err = acctest.CheckResponse(message, response, err); err != nil {
			return err
		}

		return fn(response.JSON200)
	}
}

func checkCdnDefaults(
	client cdn77.ClientWithResponsesInterface,
	cdnId *string,
	originId *string,
	cdnLabel string,
) resource.TestCheckFunc {
	return checkCdn(client, cdnId, func(c *cdn77.Cdn) error {
		sort.SliceStable(c.Cnames, func(i, j int) bool {
			return c.Cnames[i].Cname < c.Cnames[j].Cname
		})

		return errors.Join(
			acctest.NullFieldEqual("origin_id", c.OriginId, *originId),
			acctest.EqualField("label", c.Label, cdnLabel),
			acctest.NullField("note", c.Note),
			acctest.EqualField("len(cnames)", len(c.Cnames), 0),
			acctest.EqualField("cache.max_age", *c.Cache.MaxAge, cdn77.MaxAgeN17280),
			acctest.NullField("cache.max_age_404", c.Cache.MaxAge404),
			acctest.EqualField(
				"cache.requests_with_cookies_enabled",
				*c.Cache.RequestsWithCookiesEnabled,
				true,
			),
			acctest.EqualField("secure_token.type", c.SecureToken.Type, cdn77.SecureTokenTypeNone),
			acctest.EqualField("secure_token.token", c.SecureToken.Token, nil),
			acctest.EqualField(
				"query_string.ignore_type",
				c.QueryString.IgnoreType,
				cdn77.QueryStringIgnoreTypeNone,
			),
			acctest.EqualField("query_string.parameters", c.QueryString.Parameters, nil),
			acctest.EqualField("headers.cors_enabled", *c.Headers.CorsEnabled, false),
			acctest.EqualField("headers.cors_timing_enabled", *c.Headers.CorsTimingEnabled, false),
			acctest.EqualField("headers.cors_wildcard_enabled", *c.Headers.CorsWildcardEnabled, false),
			acctest.EqualField(
				"headers.host_header_forwarding_enabled",
				*c.Headers.HostHeaderForwardingEnabled,
				false,
			),
			acctest.EqualField(
				"headers.content_disposition_type",
				*c.Headers.ContentDisposition.Type,
				cdn77.ContentDispositionTypeNone,
			),
			acctest.EqualField("https_redirect.enabled", c.HttpsRedirect.Enabled, false),
			acctest.EqualField("https_redirect.code", c.HttpsRedirect.Code, nil),

			acctest.EqualField("mp4_pseudo_streaming_enabled", *c.Mp4PseudoStreaming.Enabled, false),
			acctest.EqualField("ssl.type", c.Ssl.Type, cdn77.InstantSsl),
			acctest.EqualField("ssl.ssl_id", c.Ssl.SslId, nil),
			acctest.EqualField("hotlink_protection.code", c.HotlinkProtection.Type, cdn77.Disabled),
			acctest.EqualField(
				"hotlink_protection.empty_referer_denied",
				c.HotlinkProtection.EmptyRefererDenied,
				false,
			),
			acctest.EqualField("hotlink_protection.domains", c.HotlinkProtection.Domains, nil),
			acctest.EqualField("ip_protection.type", c.IpProtection.Type, cdn77.Disabled),
			acctest.EqualField("ip_protection.ips", c.IpProtection.Ips, nil),
			acctest.EqualField("geo_protection.type", c.GeoProtection.Type, cdn77.Disabled),
			acctest.EqualField("geo_protection.countries", c.GeoProtection.Countries, nil),
			acctest.EqualField("rate_limit_enabled", c.RateLimit.Enabled, false),
			acctest.EqualField("origin_headers", c.OriginHeaders, nil),
		)
	})
}

func checkCdnsAndOriginDestroyed(client cdn77.ClientWithResponsesInterface) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		checkCdnsDestroyed(client),
		acctest.CheckOriginDestroyed(client, origin.TypeUrl),
	)
}

func checkCdnsDestroyed(client cdn77.ClientWithResponsesInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "cdn77_cdn" {
				continue
			}

			cdnId, err := strconv.Atoi(rs.Primary.Attributes["id"])
			if err != nil {
				return fmt.Errorf("unexpected CDN id: %s", rs.Primary.Attributes["id"])
			}

			response, err := client.CdnDetailWithResponse(context.Background(), cdnId)
			if err != nil {
				return fmt.Errorf("failed to fetch CDN: %w", err)
			}

			if response.JSON404 == nil {
				return errors.New("expected CDN to be deleted")
			}
		}

		return nil
	}
}

const OriginResourceConfig = `
resource "cdn77_origin_url" "url" {
	label = "origin label"
	url = "http://my-totally-random-custom-host.com"
}
`

const cdnDataSourceConfig = `
data "cdn77_cdn" "lorem" {
 id = "{id}"
}
`
