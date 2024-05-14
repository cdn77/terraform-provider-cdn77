package provider_test

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"testing"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oapi-codegen/nullable"
)

func TestAccCdnResource(t *testing.T) {
	client := acctest.GetClient(t)
	var originId string
	var cdnId int
	var cdnCreationTime string
	var cdnUrl string
	sslId := acctest.MustAddSslWithCleanup(t, client, sslTestCert1, sslTestKey)

	rsc := "cdn77_cdn.lorem"
	attrEq := func(key, value string) resource.TestCheckFunc {
		return resource.TestCheckResourceAttr(rsc, key, value)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		CheckDestroy: func(state *terraform.State) error {
			return errors.Join(checkCdnsDestroyed(client)(state), checkOriginsDestroyed(client)(state))
		},
		Steps: []resource.TestStep{
			{
				Config: OriginResourceConfig + `resource "cdn77_cdn" "lorem" {
					origin_id = cdn77_origin.url.id
					label = "my cdn"
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.url", plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(rsc, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.url", "id", func(value string) error {
						originId = value

						return acctest.NotEqual(value, "")
					}),
					resource.TestCheckResourceAttrWith(rsc, "id", func(value string) (err error) {
						cdnId, err = strconv.Atoi(value)

						return err
					}),
					attrEq("cnames.#", "0"),
					resource.TestCheckResourceAttrWith(rsc, "creation_time", func(value string) (err error) {
						cdnCreationTime = value

						return acctest.NotEqual(value, "")
					}),
					attrEq("label", "my cdn"),
					resource.TestCheckResourceAttrWith(rsc, "origin_id", func(value string) error {
						return acctest.Equal(value, originId)
					}),
					attrEq("origin_protection_enabled", "false"),
					resource.TestCheckResourceAttrWith(rsc, "url", func(value string) (err error) {
						cdnUrl = value

						return acctest.NotEqual(value, "")
					}),
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
					attrEq("waf_enabled", "false"),
					attrEq("ssl.type", string(cdn77.InstantSsl)),
					attrEq("hotlink_protection.type", string(cdn77.Disabled)),
					attrEq("hotlink_protection.empty_referer_denied", "false"),
					attrEq("ip_protection.type", string(cdn77.Disabled)),
					attrEq("geo_protection.type", string(cdn77.Disabled)),
					attrEq("rate_limit_enabled", "false"),
					attrEq("origin_headers.#", "0"),

					resource.TestCheckNoResourceAttr(rsc, "note"),
					resource.TestCheckNoResourceAttr(rsc, "cache.max_age_404"),
					resource.TestCheckNoResourceAttr(rsc, "secure_token.token"),
					resource.TestCheckNoResourceAttr(rsc, "query_string.parameters"),
					resource.TestCheckNoResourceAttr(rsc, "https_redirect.code"),
					resource.TestCheckNoResourceAttr(rsc, "ssl.ssl_id"),
					resource.TestCheckNoResourceAttr(rsc, "stream"),
					resource.TestCheckNoResourceAttr(rsc, "hotlink_protection.domains"),
					resource.TestCheckNoResourceAttr(rsc, "ip_protection.ips"),
					resource.TestCheckNoResourceAttr(rsc, "geo_protection.countries"),

					checkCdnDefaults(client, &cdnId, &originId, "my cdn"),
				),
			},
			{
				Config: OriginResourceConfig + `resource "cdn77_cdn" "lorem" {
						origin_id = cdn77_origin.url.id
						label = "changed the label"
						note = "custom note"
						cnames = ["my.cdn.cz", "other.cdn.com"]
						cache = {
							max_age = 60
							max_age_404 = 5
							requests_with_cookies_enabled = false
						}
						secure_token = {
							type = "path"
							token = "abcd1234"
						}
						query_string = {
							ignore_type = "list"
							parameters = ["param"]
						}
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
					}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(rsc, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(rsc, "id", func(value string) (err error) {
						return acctest.Equal(value, strconv.Itoa(cdnId))
					}),
					attrEq("cnames.#", "2"),
					resource.TestCheckTypeSetElemAttr(rsc, "cnames.*", "my.cdn.cz"),
					resource.TestCheckTypeSetElemAttr(rsc, "cnames.*", "other.cdn.com"),
					resource.TestCheckResourceAttrWith(rsc, "creation_time", func(value string) (err error) {
						return acctest.Equal(value, cdnCreationTime)
					}),
					attrEq("label", "changed the label"),
					attrEq("note", "custom note"),
					resource.TestCheckResourceAttrWith(rsc, "origin_id", func(value string) error {
						return acctest.Equal(value, originId)
					}),
					attrEq("origin_protection_enabled", "false"),
					resource.TestCheckResourceAttrWith(rsc, "url", func(value string) error {
						return acctest.Equal(value, cdnUrl)
					}),
					attrEq("cache.max_age", fmt.Sprintf("%d", cdn77.MaxAgeN60)),
					attrEq("cache.max_age_404", fmt.Sprintf("%d", cdn77.MaxAge404N5)),
					attrEq("cache.requests_with_cookies_enabled", "false"),
					attrEq("secure_token.type", string(cdn77.SecureTokenTypePath)),
					attrEq("secure_token.token", "abcd1234"),
					attrEq("query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeList)),
					attrEq("query_string.parameters.#", "1"),
					resource.TestCheckTypeSetElemAttr(rsc, "query_string.parameters.*", "param"),
					attrEq("headers.cors_enabled", "true"),
					attrEq("headers.cors_timing_enabled", "true"),
					attrEq("headers.cors_wildcard_enabled", "true"),
					attrEq("headers.host_header_forwarding_enabled", "true"),
					attrEq("headers.content_disposition_type", string(cdn77.ContentDispositionTypeParameter)),
					attrEq("https_redirect.enabled", "true"),
					attrEq("https_redirect.code", fmt.Sprintf("%d", cdn77.N301)),
					attrEq("mp4_pseudo_streaming_enabled", "false"),
					attrEq("waf_enabled", "false"),
					attrEq("ssl.type", string(cdn77.InstantSsl)),
					attrEq("hotlink_protection.type", string(cdn77.Disabled)),
					attrEq("hotlink_protection.empty_referer_denied", "false"),
					attrEq("ip_protection.type", string(cdn77.Disabled)),
					attrEq("geo_protection.type", string(cdn77.Disabled)),
					attrEq("rate_limit_enabled", "false"),
					attrEq("origin_headers.#", "0"),

					resource.TestCheckNoResourceAttr(rsc, "ssl.ssl_id"),
					resource.TestCheckNoResourceAttr(rsc, "stream"),
					resource.TestCheckNoResourceAttr(rsc, "hotlink_protection.domains"),
					resource.TestCheckNoResourceAttr(rsc, "ip_protection.ips"),
					resource.TestCheckNoResourceAttr(rsc, "geo_protection.countries"),

					checkCdn(client, &cdnId, func(c *cdn77.Cdn) error {
						sort.SliceStable(c.Cnames, func(i, j int) bool {
							return c.Cnames[i].Cname < c.Cnames[j].Cname
						})

						return errors.Join(
							acctest.NullFieldEqual("origin_id", c.OriginId, originId),
							acctest.EqualField("label", c.Label, "changed the label"),
							acctest.NullFieldEqual("note", c.Note, "custom note"),
							acctest.EqualField("cnames.0", c.Cnames[0].Cname, "my.cdn.cz"),
							acctest.EqualField("cnames.1", c.Cnames[1].Cname, "other.cdn.com"),
							acctest.EqualField("cache.max_age", *c.Cache.MaxAge, cdn77.MaxAgeN60),
							acctest.NullFieldEqual("cache.max_age_404", c.Cache.MaxAge404, cdn77.MaxAge404N5),
							acctest.EqualField(
								"cache.requests_with_cookies_enabled",
								*c.Cache.RequestsWithCookiesEnabled,
								false,
							),
							acctest.EqualField("secure_token.type", c.SecureToken.Type, cdn77.SecureTokenTypePath),
							acctest.EqualField("secure_token.token", *c.SecureToken.Token, "abcd1234"),
							acctest.EqualField(
								"query_string.ignore_type",
								c.QueryString.IgnoreType,
								cdn77.QueryStringIgnoreTypeList,
							),
							acctest.EqualField("query_string.parameters.0", (*c.QueryString.Parameters)[0], "param"),
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
						)
					}),
				),
			},
			{
				Config: OriginResourceConfig + acctest.Config(
					`resource "cdn77_cdn" "lorem" {
						origin_id = cdn77_origin.url.id
						label = "changed the label"
						note = "custom note"
						cnames = ["my.cdn.cz", "other.cdn.com"]
						cache = {
							max_age = 60
							max_age_404 = 5
							requests_with_cookies_enabled = false
						}
						secure_token = {
							type = "path"
							token = "abcd1234"
						}
						query_string = {
							ignore_type = "all"
						}
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
						mp4_pseudo_streaming_enabled = true
						waf_enabled = true
						ssl = {
							type = "SNI"
							ssl_id = "{sslId}"
						}
						hotlink_protection = {
							type = "blocklist"
							domains = ["example.com"] 
							empty_referer_denied = true
						}
						ip_protection = {
							type = "passlist"
							ips = ["1.1.1.1/32", "8.8.8.8/32"]
						}
						geo_protection = {
							type = "blocklist"
							countries = ["SK"]
						}
						rate_limit_enabled = true
						origin_headers = {
							abc = "v1"
							def = "v2"
						}
					}`,
					"sslId", sslId,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(rsc, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(rsc, "id", func(value string) (err error) {
						return acctest.Equal(value, strconv.Itoa(cdnId))
					}),
					attrEq("cnames.#", "2"),
					resource.TestCheckTypeSetElemAttr(rsc, "cnames.*", "my.cdn.cz"),
					resource.TestCheckTypeSetElemAttr(rsc, "cnames.*", "other.cdn.com"),
					resource.TestCheckResourceAttrWith(rsc, "creation_time", func(value string) (err error) {
						return acctest.Equal(value, cdnCreationTime)
					}),
					attrEq("label", "changed the label"),
					attrEq("note", "custom note"),
					resource.TestCheckResourceAttrWith(rsc, "origin_id", func(value string) error {
						return acctest.Equal(value, originId)
					}),
					attrEq("origin_protection_enabled", "false"),
					resource.TestCheckResourceAttrWith(rsc, "url", func(value string) error {
						return acctest.Equal(value, cdnUrl)
					}),
					attrEq("cache.max_age", fmt.Sprintf("%d", cdn77.MaxAgeN60)),
					attrEq("cache.max_age_404", fmt.Sprintf("%d", cdn77.MaxAge404N5)),
					attrEq("cache.requests_with_cookies_enabled", "false"),
					attrEq("secure_token.type", string(cdn77.SecureTokenTypePath)),
					attrEq("secure_token.token", "abcd1234"),
					attrEq("query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeAll)),
					attrEq("query_string.parameters.#", "0"),
					attrEq("headers.cors_enabled", "true"),
					attrEq("headers.cors_timing_enabled", "true"),
					attrEq("headers.cors_wildcard_enabled", "true"),
					attrEq("headers.host_header_forwarding_enabled", "true"),
					attrEq("headers.content_disposition_type", string(cdn77.ContentDispositionTypeParameter)),
					attrEq("https_redirect.enabled", "true"),
					attrEq("https_redirect.code", fmt.Sprintf("%d", cdn77.N301)),
					attrEq("mp4_pseudo_streaming_enabled", "true"),
					attrEq("waf_enabled", "true"),
					attrEq("ssl.type", string(cdn77.SNI)),
					attrEq("ssl.ssl_id", sslId),
					attrEq("hotlink_protection.type", string(cdn77.Blocklist)),
					attrEq("hotlink_protection.empty_referer_denied", "true"),
					attrEq("hotlink_protection.domains.#", "1"),
					resource.TestCheckTypeSetElemAttr(rsc, "hotlink_protection.domains.*", "example.com"),
					attrEq("ip_protection.type", string(cdn77.Passlist)),
					attrEq("ip_protection.ips.#", "2"),
					resource.TestCheckTypeSetElemAttr(rsc, "ip_protection.ips.*", "1.1.1.1/32"),
					resource.TestCheckTypeSetElemAttr(rsc, "ip_protection.ips.*", "8.8.8.8/32"),
					attrEq("geo_protection.countries.#", "1"),
					resource.TestCheckTypeSetElemAttr(rsc, "geo_protection.countries.*", "SK"),
					attrEq("geo_protection.type", string(cdn77.Blocklist)),
					attrEq("rate_limit_enabled", "true"),
					attrEq("origin_headers.%", "2"),
					attrEq("origin_headers.abc", "v1"),
					attrEq("origin_headers.def", "v2"),

					resource.TestCheckNoResourceAttr(rsc, "stream"),

					checkCdn(client, &cdnId, func(c *cdn77.Cdn) error {
						sort.SliceStable(c.Cnames, func(i, j int) bool {
							return c.Cnames[i].Cname < c.Cnames[j].Cname
						})

						sort.SliceStable(*c.IpProtection.Ips, func(i, j int) bool {
							return (*c.IpProtection.Ips)[i] < (*c.IpProtection.Ips)[j]
						})

						return errors.Join(
							acctest.NullFieldEqual("origin_id", c.OriginId, originId),
							acctest.EqualField("label", c.Label, "changed the label"),
							acctest.NullFieldEqual("note", c.Note, "custom note"),
							acctest.EqualField("cnames.0", c.Cnames[0].Cname, "my.cdn.cz"),
							acctest.EqualField("cnames.1", c.Cnames[1].Cname, "other.cdn.com"),
							acctest.EqualField("cache.max_age", *c.Cache.MaxAge, cdn77.MaxAgeN60),
							acctest.NullFieldEqual("cache.max_age_404", c.Cache.MaxAge404, cdn77.MaxAge404N5),
							acctest.EqualField(
								"cache.requests_with_cookies_enabled",
								*c.Cache.RequestsWithCookiesEnabled,
								false,
							),
							acctest.EqualField("secure_token.type", c.SecureToken.Type, cdn77.SecureTokenTypePath),
							acctest.EqualField("secure_token.token", *c.SecureToken.Token, "abcd1234"),
							acctest.EqualField(
								"query_string.ignore_type",
								c.QueryString.IgnoreType,
								cdn77.QueryStringIgnoreTypeAll,
							),
							acctest.EqualField("query_string.parameters", c.QueryString.Parameters, nil),
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
							acctest.EqualField("mp4_pseudo_streaming_enabled", *c.Mp4PseudoStreaming.Enabled, true),
							acctest.EqualField("waf_enabled", c.Waf.Enabled, true),
							acctest.EqualField("ssl.type", c.Ssl.Type, cdn77.SNI),
							acctest.EqualField("ssl.ssl_id", *c.Ssl.SslId, sslId),
							acctest.EqualField("hotlink_protection.type", c.HotlinkProtection.Type, cdn77.Blocklist),
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
							acctest.EqualField("ip_protection.type", c.IpProtection.Type, cdn77.Passlist),
							acctest.EqualField("ip_protection.ips.0", (*c.IpProtection.Ips)[0], "1.1.1.1/32"),
							acctest.EqualField("ip_protection.ips.1", (*c.IpProtection.Ips)[1], "8.8.8.8/32"),
							acctest.EqualField("geo_protection.type", c.GeoProtection.Type, cdn77.Blocklist),
							acctest.EqualField("geo_protection.countries.0", (*c.GeoProtection.Countries)[0], "SK"),
							acctest.EqualField("rate_limit_enabled", c.RateLimit.Enabled, true),
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
						)
					}),
				),
			},
			{
				Config: OriginResourceConfig + `resource "cdn77_cdn" "lorem" {
					origin_id = cdn77_origin.url.id
					label = "my cdn"
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(rsc, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(rsc, "id", func(value string) (err error) {
						return acctest.Equal(value, strconv.Itoa(cdnId))
					}),
					attrEq("cnames.#", "0"),
					resource.TestCheckResourceAttrWith(rsc, "creation_time", func(value string) (err error) {
						return acctest.Equal(value, cdnCreationTime)
					}),
					attrEq("label", "my cdn"),
					resource.TestCheckResourceAttrWith(rsc, "origin_id", func(value string) error {
						return acctest.Equal(value, originId)
					}),
					attrEq("origin_protection_enabled", "false"),
					resource.TestCheckResourceAttrWith(rsc, "url", func(value string) error {
						return acctest.Equal(value, cdnUrl)
					}),
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
					attrEq("waf_enabled", "false"),
					attrEq("ssl.type", string(cdn77.InstantSsl)),
					attrEq("hotlink_protection.type", string(cdn77.Disabled)),
					attrEq("hotlink_protection.empty_referer_denied", "false"),
					attrEq("ip_protection.type", string(cdn77.Disabled)),
					attrEq("geo_protection.type", string(cdn77.Disabled)),
					attrEq("rate_limit_enabled", "false"),
					attrEq("origin_headers.#", "0"),

					resource.TestCheckNoResourceAttr(rsc, "note"),
					resource.TestCheckNoResourceAttr(rsc, "cache.max_age_404"),
					resource.TestCheckNoResourceAttr(rsc, "secure_token.token"),
					resource.TestCheckNoResourceAttr(rsc, "query_string.parameters"),
					resource.TestCheckNoResourceAttr(rsc, "https_redirect.code"),
					resource.TestCheckNoResourceAttr(rsc, "ssl.ssl_id"),
					resource.TestCheckNoResourceAttr(rsc, "stream"),
					resource.TestCheckNoResourceAttr(rsc, "hotlink_protection.domains"),
					resource.TestCheckNoResourceAttr(rsc, "ip_protection.ips"),
					resource.TestCheckNoResourceAttr(rsc, "geo_protection.countries"),

					checkCdnDefaults(client, &cdnId, &originId, "my cdn"),
				),
			},
		},
	})
}

func checkCdn(
	client cdn77.ClientWithResponsesInterface,
	cdnId *int,
	fn func(o *cdn77.Cdn) error,
) func(*terraform.State) error {
	return func(_ *terraform.State) error {
		response, err := client.CdnDetailWithResponse(context.Background(), *cdnId)
		message := fmt.Sprintf("failed to get CDN[id=%d]: %%s", *cdnId)

		if err = acctest.CheckResponse(message, response, err); err != nil {
			return err
		}

		return fn(response.JSON200)
	}
}

func checkCdnDefaults(
	client cdn77.ClientWithResponsesInterface,
	cdnId *int,
	originId *string,
	cdnLabel string,
) func(*terraform.State) error {
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
			acctest.EqualField("waf_enabled", c.Waf.Enabled, false),
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

func checkCdnsDestroyed(client cdn77.ClientWithResponsesInterface) func(*terraform.State) error {
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
resource "cdn77_origin" "url" {
	type = "url"
	label = "origin label"
	scheme = "http"
	host = "my-totally-random-custom-host.com"
}
`
