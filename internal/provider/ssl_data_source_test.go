package provider_test

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//go:embed testdata/key.pem
var sslTestKey string

//go:embed testdata/cert1.pem
var sslTestCert1 string

//go:embed testdata/cert2.pem
var sslTestCert2 string

func init() { //nolint:gochecknoinits // here the init makes sense
	sslTestKey = strings.TrimSpace(sslTestKey)
	sslTestCert1 = strings.TrimSpace(sslTestCert1)
	sslTestCert2 = strings.TrimSpace(sslTestCert2)
}

func TestAccSslDataSource_NonExistingSsl(t *testing.T) {
	const sslId = "ae4f471f-029a-4a5c-b9bd-27ea28815de0"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      acctest.Config(sslDataSourceConfig, "id", sslId),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`SNI certificate with id "%s" couldn't be found`, sslId)),
			},
		},
	})
}

func TestAccSslDataSource(t *testing.T) {
	client := acctest.GetClient(t)
	sslId := acctest.MustAddSslWithCleanup(t, client, sslTestCert1, sslTestKey)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: acctest.Config(sslDataSourceConfig, "id", sslId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.cdn77_ssl.ipsum", "id", sslId),
					resource.TestCheckResourceAttr("data.cdn77_ssl.ipsum", "certificate", sslTestCert1),
					resource.TestCheckResourceAttr("data.cdn77_ssl.ipsum", "subjects.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.cdn77_ssl.ipsum", "subjects.*", "cdn.example.com"),
					resource.TestCheckTypeSetElemAttr("data.cdn77_ssl.ipsum", "subjects.*", "other.mycdn.cz"),
					resource.TestCheckResourceAttr("data.cdn77_ssl.ipsum", "expires_at", "2051-09-02 12:19:20"),

					resource.TestCheckNoResourceAttr("data.cdn77_ssl.ipsum", "private_key"),
				),
			},
		},
	})
}

const sslDataSourceConfig = `
data "cdn77_ssl" "ipsum" {
  id = "{id}"
}
`
