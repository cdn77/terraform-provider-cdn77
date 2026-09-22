package acctest

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

// ClusterEnvVar pins the Object Storage cluster acceptance tests create buckets on.
const ClusterEnvVar = "CDN77_TEST_OS_CLUSTER"

// ObjectStorageClusters holds the clusters the API offers; Secondary is nil when there is only one,
// which is why tests must not assume a specific cluster count, label or hostname.
type ObjectStorageClusters struct {
	Primary   cdn77.ObjectStorageCluster
	Secondary *cdn77.ObjectStorageCluster
}

func GetObjectStorageClusters(t *testing.T) ObjectStorageClusters {
	t.Helper()

	client := GetClient(t)

	response, err := client.ObjectStorageClusterListWithResponse(t.Context())
	AssertResponseOk(t, "Failed to list Object Storage clusters: %s", response, err)

	clusters := *response.JSON200
	if len(clusters) == 0 {
		t.Fatal("no Object Storage cluster available in this environment")
	}

	// the list can contain clusters that no longer accept new buckets, and nothing in the response
	// tells them apart, so CDN77_TEST_OS_CLUSTER (label or id) pins the one to use
	if wanted := os.Getenv(ClusterEnvVar); wanted != "" {
		for _, cluster := range clusters {
			if cluster.Label == wanted || cluster.Id == wanted {
				logCluster(t, "primary", cluster)

				return ObjectStorageClusters{Primary: cluster}
			}
		}

		t.Fatalf("%s=%q matches none of the clusters returned by the API", ClusterEnvVar, wanted)
	}

	// sorted so that repeated runs against the same environment pick the same primary cluster
	slices.SortFunc(clusters, func(a, b cdn77.ObjectStorageCluster) int {
		return strings.Compare(a.Label, b.Label)
	})

	result := ObjectStorageClusters{Primary: clusters[0]}
	logCluster(t, "primary", result.Primary)

	if len(clusters) > 1 {
		result.Secondary = &clusters[1]
		logCluster(t, "secondary", *result.Secondary)
	}

	return result
}

func logCluster(t *testing.T, role string, cluster cdn77.ObjectStorageCluster) {
	t.Helper()
	t.Logf("Object Storage %s cluster: %q id=%s host=%s", role, cluster.Label, cluster.Id, cluster.Host)
}

func ClusterUrl(cluster cdn77.ObjectStorageCluster) string {
	if cluster.Port == nil {
		return fmt.Sprintf("%s://%s", cluster.Scheme, cluster.Host)
	}

	return fmt.Sprintf("%s://%s:%s", cluster.Scheme, cluster.Host, strconv.Itoa(*cluster.Port))
}

func CheckClusterPort(rsc string, attr string, cluster cdn77.ObjectStorageCluster) resource.TestCheckFunc {
	if cluster.Port == nil {
		return resource.TestCheckNoResourceAttr(rsc, attr)
	}

	return resource.TestCheckResourceAttr(rsc, attr, strconv.Itoa(*cluster.Port))
}

func ClusterPortKnownValue(cluster cdn77.ObjectStorageCluster) knownvalue.Check {
	if cluster.Port == nil {
		return knownvalue.Null()
	}

	return knownvalue.Int64Exact(int64(*cluster.Port))
}
