package cdn_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/origin"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/oapi-codegen/nullable"
)

func TestAccCdnAllDataSource(t *testing.T) {
	const rsc = "data.cdn77_cdns.all"
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

	const cdn1Label = "some cdn"
	const cdn1Note = "some note"

	cdn1Request := cdn77.CdnAddJSONRequestBody{
		Label:    cdn1Label,
		OriginId: originId,
		Cnames:   util.Pointer([]string{"my.cdn.cz", "another.cdn.com"}),
		Note:     nullable.NewNullableWithValue(cdn1Note),
	}
	cdn1Response, err := client.CdnAddWithResponse(t.Context(), cdn1Request)
	acctest.AssertResponseOk(t, "Failed to create CDN: %s", cdn1Response, err)

	cdn1Id := cdn1Response.JSON201.Id
	cdn1CreationTime := cdn1Response.JSON201.CreationTime.Format(time.DateTime)
	cdn1Url := cdn1Response.JSON201.Url

	t.Cleanup(func() {
		acctest.MustDeleteCdn(t, client, cdn1Id)
	})

	const cdn2Label = "another cdn"

	cdn2Request := cdn77.CdnAddJSONRequestBody{Label: cdn2Label, OriginId: originId}
	cdn2Response, err := client.CdnAddWithResponse(t.Context(), cdn2Request)
	acctest.AssertResponseOk(t, "Failed to create CDN: %s", cdn2Response, err)

	cdn2Id := cdn2Response.JSON201.Id
	cdn2CreationTime := cdn2Response.JSON201.CreationTime.Format(time.DateTime)
	cdn2Url := cdn2Response.JSON201.Url

	t.Cleanup(func() {
		acctest.MustDeleteCdn(t, client, cdn2Id)
	})

	key := func(i int, k string) string {
		return fmt.Sprintf("cdns.%d.%s", i, k)
	}
	attrEq := func(i int, attr, value string) resource.TestCheckFunc {
		return resource.TestCheckResourceAttr(rsc, key(i, attr), value)
	}
	cdnIdAndTestCheckFnFactory := []struct {
		id      int
		factory func(i int) []resource.TestCheckFunc
	}{
		{id: cdn1Id, factory: func(i int) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				attrEq(i, "id", fmt.Sprintf("%d", cdn1Id)),
				attrEq(i, "label", "some cdn"),
				attrEq(i, "origin_id", originId),
				attrEq(i, "creation_time", cdn1CreationTime),
				attrEq(i, "url", cdn1Url),
				attrEq(i, "cache.max_age", fmt.Sprintf("%d", cdn77.MaxAgeN17280)),
				attrEq(i, "cache.requests_with_cookies_enabled", "true"),
				attrEq(i, "cnames.#", "2"),
				resource.TestCheckTypeSetElemAttr(rsc, key(i, "cnames.*"), "my.cdn.cz"),
				resource.TestCheckTypeSetElemAttr(rsc, key(i, "cnames.*"), "another.cdn.com"),
				attrEq(i, "geo_protection.type", string(cdn77.Disabled)),
				attrEq(i, "headers.content_disposition_type", string(cdn77.ContentDispositionTypeNone)),
				attrEq(i, "headers.cors_enabled", "false"),
				attrEq(i, "headers.cors_timing_enabled", "false"),
				attrEq(i, "headers.cors_wildcard_enabled", "false"),
				attrEq(i, "headers.host_header_forwarding_enabled", "false"),
				attrEq(i, "hotlink_protection.empty_referer_denied", "false"),
				attrEq(i, "hotlink_protection.type", string(cdn77.Disabled)),
				attrEq(i, "https_redirect.enabled", "false"),
				attrEq(i, "ip_protection.type", string(cdn77.Disabled)),
				attrEq(i, "mp4_pseudo_streaming_enabled", "false"),
				attrEq(i, "note", cdn1Note),
				attrEq(i, "origin_headers.#", "0"),
				attrEq(i, "query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeNone)),
				attrEq(i, "rate_limit_enabled", "false"),
				attrEq(i, "secure_token.type", string(cdn77.SecureTokenTypeNone)),
				attrEq(i, "ssl.type", string(cdn77.InstantSsl)),

				resource.TestCheckNoResourceAttr(rsc, key(i, "stream")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "cache.max_age_404")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "geo_protection.countries")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "hotlink_protection.domains")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "https_redirect.code")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "ip_protection.ips")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "query_string.parameters")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "secure_token.token")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "ssl.ssl_id")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "waf_enabled")),
			}
		}},
		{id: cdn2Id, factory: func(i int) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				attrEq(i, "id", fmt.Sprintf("%d", cdn2Id)),
				attrEq(i, "label", "another cdn"),
				attrEq(i, "origin_id", originId),
				attrEq(i, "creation_time", cdn2CreationTime),
				attrEq(i, "url", cdn2Url),
				attrEq(i, "cache.max_age", fmt.Sprintf("%d", cdn77.MaxAgeN17280)),
				attrEq(i, "cache.requests_with_cookies_enabled", "true"),
				attrEq(i, "cnames.#", "0"),
				attrEq(i, "geo_protection.type", string(cdn77.Disabled)),
				attrEq(i, "headers.content_disposition_type", string(cdn77.ContentDispositionTypeNone)),
				attrEq(i, "headers.cors_enabled", "false"),
				attrEq(i, "headers.cors_timing_enabled", "false"),
				attrEq(i, "headers.cors_wildcard_enabled", "false"),
				attrEq(i, "headers.host_header_forwarding_enabled", "false"),
				attrEq(i, "hotlink_protection.empty_referer_denied", "false"),
				attrEq(i, "hotlink_protection.type", string(cdn77.Disabled)),
				attrEq(i, "https_redirect.enabled", "false"),
				attrEq(i, "ip_protection.type", string(cdn77.Disabled)),
				attrEq(i, "mp4_pseudo_streaming_enabled", "false"),
				attrEq(i, "origin_headers.#", "0"),
				attrEq(i, "query_string.ignore_type", string(cdn77.QueryStringIgnoreTypeNone)),
				attrEq(i, "rate_limit_enabled", "false"),
				attrEq(i, "secure_token.type", string(cdn77.SecureTokenTypeNone)),
				attrEq(i, "ssl.type", string(cdn77.InstantSsl)),

				resource.TestCheckNoResourceAttr(rsc, key(i, "stream")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "cache.max_age_404")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "geo_protection.countries")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "hotlink_protection.domains")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "https_redirect.code")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "ip_protection.ips")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "note")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "query_string.parameters")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "secure_token.token")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "ssl.ssl_id")),
			}
		}},
	}

	sort.SliceStable(cdnIdAndTestCheckFnFactory, func(i, j int) bool {
		return cdnIdAndTestCheckFnFactory[i].id < cdnIdAndTestCheckFnFactory[j].id
	})

	testCheckFns := []resource.TestCheckFunc{resource.TestCheckResourceAttr(rsc, "cdns.#", "2")}

	for i, x := range cdnIdAndTestCheckFnFactory {
		testCheckFns = append(testCheckFns, x.factory(i)...)
	}

	acctest.Run(t, nil, resource.TestStep{
		Config: cdnsDataSourceConfig,
		Check:  resource.ComposeAggregateTestCheckFunc(testCheckFns...),
	})
}

func TestAccCdnAllDataSource_Empty(t *testing.T) {
	acctest.Run(t, nil, resource.TestStep{
		Config: cdnsDataSourceConfig,
		Check:  resource.TestCheckResourceAttr("data.cdn77_cdns.all", "cdns.#", "0"),
	})
}

const cdnsDataSourceConfig = `
data "cdn77_cdns" "all" {
}
`
