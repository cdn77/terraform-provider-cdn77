package provider_test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/oapi-codegen/nullable"
)

func TestAccCdnsDataSource_All(t *testing.T) {
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

	const cdn1Label = "some cdn"
	const cdn1Note = "some note"

	cdn1Request := cdn77.CdnAddJSONRequestBody{
		Cnames:   util.Pointer([]string{"my.cdn.cz", "another.cdn.com"}),
		Label:    cdn1Label,
		Note:     nullable.NewNullableWithValue(cdn1Note),
		OriginId: originId,
	}
	cdn1Response, err := client.CdnAddWithResponse(context.Background(), cdn1Request)
	acctest.AssertResponseOk(t, "Failed to create CDN: %s", cdn1Response, err)

	cdn1Id := cdn1Response.JSON201.Id
	cdn1CreationTime := cdn1Response.JSON201.CreationTime.Format(time.DateTime)
	cdn1Url := cdn1Response.JSON201.Url

	t.Cleanup(func() {
		acctest.MustDeleteCdn(t, client, cdn1Id)
	})

	const cdn2Label = "another cdn"

	cdn2Request := cdn77.CdnAddJSONRequestBody{Label: cdn2Label, OriginId: originId}
	cdn2Response, err := client.CdnAddWithResponse(context.Background(), cdn2Request)
	acctest.AssertResponseOk(t, "Failed to create CDN: %s", cdn2Response, err)

	cdn2Id := cdn2Response.JSON201.Id
	cdn2CreationTime := cdn2Response.JSON201.CreationTime.Format(time.DateTime)
	cdn2Url := cdn2Response.JSON201.Url

	t.Cleanup(func() {
		acctest.MustDeleteCdn(t, client, cdn2Id)
	})

	rsc := "data.cdn77_cdns.all"
	key := func(i int, k string) string {
		return fmt.Sprintf("cdns.%d.%s", i, k)
	}
	cdnIdAndTestCheckFnFactory := []struct {
		id      int
		factory func(i int) []resource.TestCheckFunc
	}{
		{id: cdn1Id, factory: func(i int) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(rsc, key(i, "id"), fmt.Sprintf("%d", cdn1Id)),
				resource.TestCheckResourceAttr(rsc, key(i, "cnames.#"), "2"),
				resource.TestCheckTypeSetElemAttr(rsc, key(i, "cnames.*"), "my.cdn.cz"),
				resource.TestCheckTypeSetElemAttr(rsc, key(i, "cnames.*"), "another.cdn.com"),
				resource.TestCheckResourceAttr(rsc, key(i, "creation_time"), cdn1CreationTime),
				resource.TestCheckResourceAttr(rsc, key(i, "label"), cdn1Label),
				resource.TestCheckResourceAttr(rsc, key(i, "note"), cdn1Note),
				resource.TestCheckResourceAttr(rsc, key(i, "origin_id"), originId),
				resource.TestCheckResourceAttr(rsc, key(i, "url"), cdn1Url),
				resource.TestCheckResourceAttr(rsc, key(i, "mp4_pseudo_streaming_enabled"), "false"),
			}
		}},
		{id: cdn2Id, factory: func(i int) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(rsc, key(i, "id"), fmt.Sprintf("%d", cdn2Id)),
				resource.TestCheckResourceAttr(rsc, key(i, "cnames.#"), "0"),
				resource.TestCheckResourceAttr(rsc, key(i, "creation_time"), cdn2CreationTime),
				resource.TestCheckResourceAttr(rsc, key(i, "label"), cdn2Label),
				resource.TestCheckResourceAttr(rsc, key(i, "origin_id"), originId),
				resource.TestCheckResourceAttr(rsc, key(i, "url"), cdn2Url),
				resource.TestCheckResourceAttr(rsc, key(i, "mp4_pseudo_streaming_enabled"), "false"),

				resource.TestCheckNoResourceAttr(rsc, key(i, "note")),
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

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: cdnsDataSourceConfig,
				Check:  resource.ComposeAggregateTestCheckFunc(testCheckFns...),
			},
		},
	})
}

func TestAccCdnsDataSource_Empty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: cdnsDataSourceConfig,
				Check:  resource.TestCheckResourceAttr("data.cdn77_cdns.all", "cdns.#", "0"),
			},
		},
	})
}

const cdnsDataSourceConfig = `
data "cdn77_cdns" "all" {
}
`
