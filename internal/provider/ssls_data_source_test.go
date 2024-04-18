package provider_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSslsDataSource_All(t *testing.T) {
	client := acctest.GetClient(t)

	ssl1Id := acctest.MustAddSslWithCleanup(t, client, sslTestCert1, sslTestKey)
	ssl2Id := acctest.MustAddSslWithCleanup(t, client, sslTestCert2, sslTestKey)

	rsc := "data.cdn77_ssls.all"
	key := func(i int, k string) string {
		return fmt.Sprintf("ssls.%d.%s", i, k)
	}
	sslIdAndTestCheckFnFactory := []struct {
		id      string
		factory func(i int) []resource.TestCheckFunc
	}{
		{id: ssl1Id, factory: func(i int) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(rsc, key(i, "id"), ssl1Id),
				resource.TestCheckResourceAttr(rsc, key(i, "certificate"), sslTestCert1),
				resource.TestCheckResourceAttr(rsc, key(i, "subjects.#"), "2"),
				resource.TestCheckTypeSetElemAttr(rsc, key(i, "subjects.*"), "cdn.example.com"),
				resource.TestCheckTypeSetElemAttr(rsc, key(i, "subjects.*"), "other.mycdn.cz"),
				resource.TestCheckResourceAttr(rsc, key(i, "expires_at"), "2051-09-02 12:19:20"),

				resource.TestCheckNoResourceAttr(rsc, key(i, "private_key")),
			}
		}},
		{id: ssl2Id, factory: func(i int) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(rsc, key(i, "id"), ssl2Id),
				resource.TestCheckResourceAttr(rsc, key(i, "certificate"), sslTestCert2),
				resource.TestCheckResourceAttr(rsc, key(i, "subjects.#"), "1"),
				resource.TestCheckTypeSetElemAttr(rsc, key(i, "subjects.*"), "mycdn.cz"),
				resource.TestCheckResourceAttr(rsc, key(i, "expires_at"), "2051-09-03 07:51:29"),

				resource.TestCheckNoResourceAttr(rsc, key(i, "private_key")),
			}
		}},
	}

	sort.SliceStable(sslIdAndTestCheckFnFactory, func(i, j int) bool {
		return sslIdAndTestCheckFnFactory[i].id < sslIdAndTestCheckFnFactory[j].id
	})

	testCheckFns := []resource.TestCheckFunc{resource.TestCheckResourceAttr(rsc, "ssls.#", "2")}

	for i, x := range sslIdAndTestCheckFnFactory {
		testCheckFns = append(testCheckFns, x.factory(i)...)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: sslsDataSourceConfig,
				Check:  resource.ComposeAggregateTestCheckFunc(testCheckFns...),
			},
		},
	})
}

func TestAccSslsDataSource_Empty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: sslsDataSourceConfig,
				Check:  resource.TestCheckResourceAttr("data.cdn77_ssls.all", "ssls.#", "0"),
			},
		},
	})
}

const sslsDataSourceConfig = `
data "cdn77_ssls" "all" {
}
`
