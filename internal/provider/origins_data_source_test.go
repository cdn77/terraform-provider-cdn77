package provider_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/oapi-codegen/nullable"
)

func TestAccOriginsDataSource_All(t *testing.T) {
	client := acctest.GetClient(t)

	const origin1Host = "my-totally-random-custom-host.com"
	const origin1Label = "random origin"
	const origin1Note = "some note"
	const origin1Scheme = "https"

	request1 := cdn77.OriginAddUrlJSONRequestBody{
		Host:   origin1Host,
		Label:  origin1Label,
		Note:   nullable.NewNullableWithValue(origin1Note),
		Scheme: origin1Scheme,
	}
	response1, err := client.OriginAddUrlWithResponse(context.Background(), request1)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response1, err)

	origin1Id := response1.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, provider.OriginTypeUrl, origin1Id)
	})

	const origin2AwsKeyId = "someKeyId"
	const origin2AwsKeySecret = "someKeySecret"
	const origin2AwsRegion = "eu"
	const origin2BaseDir = "some-dir"
	const origin2Host = "some-other-totally-random-custom-host.com"
	const origin2Label = "another origin"
	const origin2Scheme = "http"

	request2 := cdn77.OriginAddAwsJSONRequestBody{
		AwsAccessKeyId:     nullable.NewNullableWithValue(origin2AwsKeyId),
		AwsAccessKeySecret: nullable.NewNullableWithValue(origin2AwsKeySecret),
		AwsRegion:          nullable.NewNullableWithValue(origin2AwsRegion),
		BaseDir:            nullable.NewNullableWithValue(origin2BaseDir),
		Host:               origin2Host,
		Label:              origin2Label,
		Scheme:             origin2Scheme,
	}
	response2, err := client.OriginAddAwsWithResponse(context.Background(), request2)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response2, err)

	origin2Id := response2.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, provider.OriginTypeAws, origin2Id)
	})

	rsc := "data.cdn77_origins.all"
	key := func(i int, k string) string {
		return fmt.Sprintf("origins.%d.%s", i, k)
	}

	originIdAndTestCheckFnFactory := []struct {
		id      string
		factory func(i int) []resource.TestCheckFunc
	}{
		{id: origin1Id, factory: func(i int) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(rsc, key(i, "id"), origin1Id),
				resource.TestCheckResourceAttr(rsc, key(i, "type"), provider.OriginTypeUrl),
				resource.TestCheckResourceAttr(rsc, key(i, "label"), origin1Label),
				resource.TestCheckResourceAttr(rsc, key(i, "note"), origin1Note),
				resource.TestCheckResourceAttr(rsc, key(i, "scheme"), origin1Scheme),
				resource.TestCheckResourceAttr(rsc, key(i, "host"), origin1Host),
				resource.TestCheckNoResourceAttr(rsc, key(i, "aws_access_key_id")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "aws_access_key_secret")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "aws_region")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "port")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "base_dir")),
			}
		}},
		{id: origin2Id, factory: func(i int) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(rsc, key(i, "id"), origin2Id),
				resource.TestCheckResourceAttr(rsc, key(i, "type"), provider.OriginTypeAws),
				resource.TestCheckResourceAttr(rsc, key(i, "label"), origin2Label),
				resource.TestCheckResourceAttr(rsc, key(i, "aws_access_key_id"), origin2AwsKeyId),
				resource.TestCheckResourceAttr(rsc, key(i, "aws_region"), origin2AwsRegion),
				resource.TestCheckResourceAttr(rsc, key(i, "scheme"), origin2Scheme),
				resource.TestCheckResourceAttr(rsc, key(i, "host"), origin2Host),
				resource.TestCheckResourceAttr(rsc, key(i, "base_dir"), origin2BaseDir),
				resource.TestCheckNoResourceAttr(rsc, key(i, "note")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "aws_access_key_secret")),
				resource.TestCheckNoResourceAttr(rsc, key(i, "port")),
			}
		}},
	}

	sort.SliceStable(originIdAndTestCheckFnFactory, func(i, j int) bool {
		return originIdAndTestCheckFnFactory[i].id < originIdAndTestCheckFnFactory[j].id
	})

	testCheckFns := []resource.TestCheckFunc{resource.TestCheckResourceAttr(rsc, "origins.#", "2")}

	for i, x := range originIdAndTestCheckFnFactory {
		testCheckFns = append(testCheckFns, x.factory(i)...)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: originsDataSourceConfig,
				Check:  resource.ComposeAggregateTestCheckFunc(testCheckFns...),
			},
		},
	})
}

func TestAccOriginsDataSource_Empty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: originsDataSourceConfig,
				Check:  resource.TestCheckResourceAttr("data.cdn77_origins.all", "origins.#", "0"),
			},
		},
	})
}

const originsDataSourceConfig = `
data "cdn77_origins" "all" {
}
`
