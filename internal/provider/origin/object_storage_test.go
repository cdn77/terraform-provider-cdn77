package origin_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/origin"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/shared"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oapi-codegen/nullable"
)

func TestAccOrigin_ObjectStorageResource(t *testing.T) {
	const rsc = "cdn77_origin_object_storage.os"
	client := acctest.GetClient(t)
	clusters := acctest.GetObjectStorageClusters(t)
	bucketName := "my-bucket-" + uuid.New().String()
	anotherBucketName := "my-bucket-" + uuid.New().String()
	var originId string

	config := func(cluster cdn77.ObjectStorageCluster, bucket string, extraAttrs string) string {
		return acctest.Config(`resource "cdn77_origin_object_storage" "os" {
				label = "{bucket}"
				cluster_id = "{clusterId}"
				bucket_name = "{bucket}"
				{extraAttrs}
			}`,
			"bucket", bucket,
			"clusterId", cluster.Id,
			"extraAttrs", extraAttrs,
		)
	}

	checks := func(cluster cdn77.ObjectStorageCluster, bucket string, note *string) []resource.TestCheckFunc {
		attrs := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(rsc, "label", bucket),
			resource.TestCheckResourceAttr(rsc, "url", acctest.ClusterUrl(cluster)),
			resource.TestCheckResourceAttr(rsc, "url_parts.scheme", cluster.Scheme),
			resource.TestCheckResourceAttr(rsc, "url_parts.host", cluster.Host),
			acctest.CheckClusterPort(rsc, "url_parts.port", cluster),
			resource.TestCheckResourceAttr(rsc, "bucket_name", bucket),
			resource.TestCheckResourceAttr(rsc, "cluster_id", cluster.Id),
			resource.TestCheckResourceAttr(rsc, "usage.files", "0"),
			resource.TestCheckResourceAttr(rsc, "usage.size_bytes", "0"),
			resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
			checkObjectStorage(client, &originId, func(o *cdn77.ObjectStorageOriginDetail) error {
				noteCheck := acctest.NullField("note", o.Note)
				if note != nil {
					noteCheck = acctest.NullFieldEqual("note", o.Note, *note)
				}

				return errors.Join(
					acctest.EqualField("type", o.Type, origin.TypeObjectStorage),
					acctest.EqualField("label", o.Label, bucket),
					noteCheck,
					acctest.EqualField("scheme", o.Scheme, cdn77.OriginScheme(cluster.Scheme)),
					acctest.EqualField("host", o.Host, cluster.Host),
					acctest.EqualField("bucket_name", o.BucketName, bucket),
				)
			}),
		}

		if note != nil {
			attrs = append(attrs, resource.TestCheckResourceAttr(rsc, "note", *note))
		} else {
			attrs = append(attrs, resource.TestCheckNoResourceAttr(rsc, "note"))
		}

		return attrs
	}

	note := "some note"
	steps := []resource.TestStep{
		{
			Config:           config(clusters.Primary, bucketName, ""),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionCreate),
			Check: resource.ComposeAggregateTestCheckFunc(
				append([]resource.TestCheckFunc{acctest.CheckAndAssignAttr(rsc, "id", &originId)},
					checks(clusters.Primary, bucketName, nil)...)...,
			),
		},
		{
			Config:           config(clusters.Primary, bucketName, `note = "some note"`),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionUpdate),
			Check: resource.ComposeAggregateTestCheckFunc(
				append([]resource.TestCheckFunc{acctest.CheckAttr(rsc, "id", &originId)},
					checks(clusters.Primary, bucketName, &note)...)...,
			),
		},
		{
			// "acl" is deprecated and inert, so changing it must no longer replace the resource.
			Config:           config(clusters.Primary, bucketName, "note = \"some note\"\nacl = \"authenticated-read\""),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionUpdate),
			Check: resource.ComposeAggregateTestCheckFunc(
				append([]resource.TestCheckFunc{
					acctest.CheckAttr(rsc, "id", &originId),
					resource.TestCheckResourceAttr(rsc, "acl", "authenticated-read"),
				}, checks(clusters.Primary, bucketName, &note)...)...,
			),
		},
		{
			Config:           config(clusters.Primary, anotherBucketName, `note = "some note"`),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionDestroyBeforeCreate),
			Check: resource.ComposeAggregateTestCheckFunc(
				append([]resource.TestCheckFunc{acctest.CheckAndReassignAttr(rsc, "id", &originId)},
					checks(clusters.Primary, anotherBucketName, &note)...)...,
			),
		},
	}

	// last, so that skipping it on single-cluster environments leaves the other steps untouched
	if clusters.Secondary != nil {
		steps = append(steps, resource.TestStep{
			Config:           config(*clusters.Secondary, anotherBucketName, `note = "some note"`),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionDestroyBeforeCreate),
			Check: resource.ComposeAggregateTestCheckFunc(
				append([]resource.TestCheckFunc{acctest.CheckAndReassignAttr(rsc, "id", &originId)},
					checks(*clusters.Secondary, anotherBucketName, &note)...)...,
			),
		})
	} else {
		t.Log("only one Object Storage cluster available; skipping the cluster change step")
	}

	acctest.Run(t, acctest.CheckOriginDestroyed(client, origin.TypeObjectStorage), steps...)
}

func TestAccOrigin_ObjectStorageResource_Import(t *testing.T) {
	const rsc = "cdn77_origin_object_storage.os"
	client := acctest.GetClient(t)
	clusters := acctest.GetObjectStorageClusters(t)
	bucketName := "my-bucket-" + uuid.New().String()
	var originId, clusterId string

	acctest.Run(t, acctest.CheckOriginDestroyed(client, origin.TypeObjectStorage),
		resource.TestStep{
			Config: acctest.Config(`resource "cdn77_origin_object_storage" "os" {
					label = "{bucketName}"
					note = "some note"
					acl = "private"
					cluster_id = "{clusterId}"
					bucket_name = "{bucketName}"
				}`, "bucketName", bucketName, "clusterId", clusters.Primary.Id),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAndAssignAttr(rsc, "id", &originId),
				acctest.CheckAndAssignAttr(rsc, "cluster_id", &clusterId),
			),
		},
		resource.TestStep{
			ResourceName: rsc,
			ImportState:  true,
			ImportStateIdFunc: func(*terraform.State) (string, error) {
				return fmt.Sprintf("%s,private,%s", originId, clusterId), nil
			},
			ImportStateVerify: true,
		},
		resource.TestStep{
			ResourceName: rsc,
			ImportState:  true,
			ImportStateIdFunc: func(*terraform.State) (string, error) {
				return fmt.Sprintf("%s,%s", originId, clusterId), nil
			},
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"acl"},
		},
	)
}

func TestAccOrigin_ObjectStorageDataSource_OnlyRequiredFields(t *testing.T) {
	const nonExistingOriginId = "bcd7b5bb-a044-4611-82e4-3f3b2a3cda13"
	const rsc = "data.cdn77_origin_object_storage.os"
	client := acctest.GetClient(t)
	originBucketName := "my-bucket-" + uuid.New().String()
	label := originBucketName
	request := cdn77.OriginCreateObjectStorageJSONRequestBody{
		BucketName: originBucketName,
		ClusterId:  acctest.GetObjectStorageClusters(t).Primary.Id,
		Label:      label,
	}

	response, err := client.OriginCreateObjectStorageWithResponse(t.Context(), request)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response, err)

	originId := response.JSON201.Id
	scheme := string(response.JSON201.Scheme)
	host := response.JSON201.Host
	port := response.JSON201.Port
	urlModel := shared.NewUrlModel(t.Context(), scheme, host, port, nullable.NewNullNullable[string]())
	originUrl := urlModel.Url.ValueString()

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, origin.TypeObjectStorage, originId)
	})

	acctest.Run(t, nil,
		resource.TestStep{
			Config:      acctest.Config(objectStorageDataSourceConfig, "id", nonExistingOriginId),
			ExpectError: regexp.MustCompile(fmt.Sprintf(`.*?"%s".*?not found.*?`, nonExistingOriginId)),
		},
		resource.TestStep{
			Config: acctest.Config(objectStorageDataSourceConfig, "id", originId),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(rsc, "id", originId),
				resource.TestCheckResourceAttr(rsc, "label", label),
				resource.TestCheckResourceAttr(rsc, "url", originUrl),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", scheme),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", host),
				resource.TestCheckResourceAttr(rsc, "url_parts.port", "443"),
				func(state *terraform.State) error {
					if port, err := port.Get(); err == nil {
						return resource.TestCheckResourceAttr(rsc, "url_parts.port", strconv.Itoa(port))(state)
					}

					return resource.TestCheckNoResourceAttr(rsc, "url_parts.port")(state)
				},
				resource.TestCheckResourceAttr(rsc, "bucket_name", originBucketName),
				resource.TestCheckNoResourceAttr(rsc, "note"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
				resource.TestCheckNoResourceAttr(rsc, "acl"),
				resource.TestCheckNoResourceAttr(rsc, "cluster_id"),
				resource.TestCheckNoResourceAttr(rsc, "access_key_id"),
				resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
			),
		},
	)
}

func TestAccOrigin_ObjectStorageDataSource_AllFields(t *testing.T) {
	const rsc = "data.cdn77_origin_object_storage.os"
	const note = "some note"
	client := acctest.GetClient(t)
	originBucketName := "my-bucket-" + uuid.New().String()
	label := originBucketName
	request := cdn77.OriginCreateObjectStorageJSONRequestBody{
		BucketName: originBucketName,
		ClusterId:  acctest.GetObjectStorageClusters(t).Primary.Id,
		Label:      label,
		Note:       nullable.NewNullableWithValue(note),
	}

	response, err := client.OriginCreateObjectStorageWithResponse(t.Context(), request)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response, err)

	originId := response.JSON201.Id
	scheme := string(response.JSON201.Scheme)
	host := response.JSON201.Host
	port := response.JSON201.Port
	urlModel := shared.NewUrlModel(t.Context(), scheme, host, port, nullable.NewNullNullable[string]())
	originUrl := urlModel.Url.ValueString()

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, origin.TypeObjectStorage, originId)
	})

	acctest.Run(t, nil, resource.TestStep{
		Config: acctest.Config(objectStorageDataSourceConfig, "id", originId),
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(rsc, "id", originId),
			resource.TestCheckResourceAttr(rsc, "label", label),
			resource.TestCheckResourceAttr(rsc, "note", note),
			resource.TestCheckResourceAttr(rsc, "url", originUrl),
			resource.TestCheckResourceAttr(rsc, "url_parts.scheme", scheme),
			resource.TestCheckResourceAttr(rsc, "url_parts.host", host),
			func(state *terraform.State) error {
				if port, err := port.Get(); err == nil {
					return resource.TestCheckResourceAttr(rsc, "url_parts.port", strconv.Itoa(port))(state)
				}

				return resource.TestCheckNoResourceAttr(rsc, "url_parts.port")(state)
			},
			resource.TestCheckResourceAttr(rsc, "bucket_name", originBucketName),
			resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
			resource.TestCheckNoResourceAttr(rsc, "acl"),
			resource.TestCheckNoResourceAttr(rsc, "cluster_id"),
			resource.TestCheckNoResourceAttr(rsc, "access_key_id"),
			resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
		),
	})
}

func checkObjectStorage(
	client cdn77.ClientWithResponsesInterface,
	originId *string,
	fn func(o *cdn77.ObjectStorageOriginDetail) error,
) func(*terraform.State) error {
	return func(*terraform.State) error {
		response, err := client.OriginDetailObjectStorageWithResponse(context.Background(), *originId)
		message := fmt.Sprintf("failed to get Origin[id=%s]: %%s", *originId)

		if err = acctest.CheckResponse(message, response, err); err != nil {
			return err
		}

		return fn(response.JSON200)
	}
}

const objectStorageDataSourceConfig = `
data "cdn77_origin_object_storage" "os" {
  id = "{id}"
}
`
