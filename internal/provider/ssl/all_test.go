package ssl_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest/testdata"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSslAllDataSource(t *testing.T) {
	const rsc = "data.cdn77_ssls.all"
	client := acctest.GetClient(t)
	ssl1Id := acctest.MustAddSslWithCleanup(t, client, testdata.SslCert1, testdata.SslKey)
	ssl2Id := acctest.MustAddSslWithCleanup(t, client, testdata.SslCert2, testdata.SslKey)

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
				resource.TestCheckResourceAttr(rsc, key(i, "certificate"), testdata.SslCert1),
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
				resource.TestCheckResourceAttr(rsc, key(i, "certificate"), testdata.SslCert2),
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

	testCheckFns := make([]resource.TestCheckFunc, 0, 20)
	testCheckFns = append(testCheckFns, resource.TestCheckResourceAttr(rsc, "ssls.#", "2"))

	for i, x := range sslIdAndTestCheckFnFactory {
		testCheckFns = append(testCheckFns, x.factory(i)...)
	}

	acctest.Run(t, nil, resource.TestStep{
		Config: sslsDataSourceConfig,
		Check:  resource.ComposeAggregateTestCheckFunc(testCheckFns...),
	})
}

func TestAccSslAllDataSource_Empty(t *testing.T) {
	acctest.Run(t, nil, resource.TestStep{
		Config: sslsDataSourceConfig,
		Check:  resource.TestCheckResourceAttr("data.cdn77_ssls.all", "ssls.#", "0"),
	})
}

const sslsDataSourceConfig = `
data "cdn77_ssls" "all" {
}
`
