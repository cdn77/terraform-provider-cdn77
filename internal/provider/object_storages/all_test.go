package object_storages_test

import (
	"regexp"
	"testing"

	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccObjectStoragesDataSource(t *testing.T) {
	acctest.Run(t, nil, resource.TestStep{
		Config: objectStoragesDataSourceConfig,
		ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("data.cdn77_object_storages.all",
				tfjsonpath.New("clusters"),
				knownvalue.SetPartial([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						"id": knownvalue.StringRegexp(regexp.MustCompile(
							`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
						)),
						"label": knownvalue.StringExact("EU"),
						"url": knownvalue.StringRegexp(regexp.MustCompile(
							`^https://([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}(:\d+)?$`,
						)),
						"url_parts": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"scheme": knownvalue.StringExact("https"),
							"host": knownvalue.StringRegexp(regexp.MustCompile(
								`^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`,
							)),
							"port":      knownvalue.Int64Exact(443),
							"base_path": knownvalue.Null(),
						}),
					}),
				}),
			),
		},
	})
}

const objectStoragesDataSourceConfig = `
data "cdn77_object_storages" "all" {
}

locals {
	eu_cluster_id = one([for os in data.cdn77_object_storages.all.clusters : os.id if os.label == "EU"])
	us_cluster_id = one([for os in data.cdn77_object_storages.all.clusters : os.id if os.label == "US"])
}
`
