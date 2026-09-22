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
	cluster := acctest.GetObjectStorageClusters(t).Primary

	acctest.Run(t, nil, resource.TestStep{
		Config: objectStoragesDataSourceConfig,
		ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("data.cdn77_object_storages.all",
				tfjsonpath.New("clusters"),
				knownvalue.SetPartial([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						"id":    knownvalue.StringExact(cluster.Id),
						"label": knownvalue.StringExact(cluster.Label),
						"url":   knownvalue.StringExact(acctest.ClusterUrl(cluster)),
						"url_parts": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"scheme": knownvalue.StringExact(cluster.Scheme),
							"host": knownvalue.StringRegexp(regexp.MustCompile(
								`^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`,
							)),
							"port":      acctest.ClusterPortKnownValue(cluster),
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
`
