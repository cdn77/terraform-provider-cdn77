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
	bucketName := "my-bucket-" + uuid.New().String()
	anotherBucketName := "my-bucket-" + uuid.New().String()
	var originId string
	var clusterId string

	acctest.Run(t, acctest.CheckOriginDestroyed(client, origin.TypeObjectStorage),
		resource.TestStep{
			Config: acctest.Config(objectStoragesDataSourceConfig+`resource "cdn77_origin_object_storage" "os" {
					label = "{bucketName}"
					acl = "private"
					cluster_id = local.eu_cluster_id
					bucket_name = "{bucketName}"
				}`, "bucketName", bucketName),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionCreate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAndAssignAttr(rsc, "id", &originId),
				resource.TestCheckResourceAttr(rsc, "label", bucketName),
				resource.TestCheckResourceAttr(rsc, "url", "https://eu-1.cdn77-storage.com:443"),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", "https"),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", "eu-1.cdn77-storage.com"),
				resource.TestCheckResourceAttr(rsc, "url_parts.port", "443"),
				resource.TestCheckResourceAttr(rsc, "bucket_name", bucketName),
				resource.TestCheckResourceAttr(rsc, "acl", "private"),
				acctest.CheckAndAssignAttr(rsc, "cluster_id", &clusterId),
				resource.TestCheckResourceAttr(rsc, "usage.files", "0"),
				resource.TestCheckResourceAttr(rsc, "usage.size_bytes", "0"),
				resource.TestCheckNoResourceAttr(rsc, "note"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
				checkObjectStorage(client, &originId, func(o *cdn77.ObjectStorageOriginDetail) error {
					return errors.Join(
						acctest.EqualField("type", o.Type, origin.TypeObjectStorage),
						acctest.EqualField("label", o.Label, bucketName),
						acctest.NullField("note", o.Note),
						acctest.EqualField("scheme", o.Scheme, "https"),
						acctest.EqualField("host", o.Host, "eu-1.cdn77-storage.com"),
						acctest.NullFieldEqual("port", o.Port, 443),
						acctest.EqualField("bucket_name", o.BucketName, bucketName),
					)
				}),
			),
		},
		resource.TestStep{
			Config: acctest.Config(objectStoragesDataSourceConfig+`resource "cdn77_origin_object_storage" "os" {
					label = "{bucketName}"
					note = "some note"
					acl = "private"
					cluster_id = local.eu_cluster_id
					bucket_name = "{bucketName}"
				}`, "bucketName", bucketName),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionUpdate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAttr(rsc, "id", &originId),
				resource.TestCheckResourceAttr(rsc, "label", bucketName),
				resource.TestCheckResourceAttr(rsc, "note", "some note"),
				resource.TestCheckResourceAttr(rsc, "url", "https://eu-1.cdn77-storage.com:443"),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", "https"),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", "eu-1.cdn77-storage.com"),
				resource.TestCheckResourceAttr(rsc, "url_parts.port", "443"),
				resource.TestCheckResourceAttr(rsc, "bucket_name", bucketName),
				resource.TestCheckResourceAttr(rsc, "acl", "private"),
				acctest.CheckAttr(rsc, "cluster_id", &clusterId),
				resource.TestCheckResourceAttr(rsc, "usage.files", "0"),
				resource.TestCheckResourceAttr(rsc, "usage.size_bytes", "0"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
				checkObjectStorage(client, &originId, func(o *cdn77.ObjectStorageOriginDetail) error {
					return errors.Join(
						acctest.EqualField("type", o.Type, origin.TypeObjectStorage),
						acctest.EqualField("label", o.Label, bucketName),
						acctest.NullFieldEqual("note", o.Note, "some note"),
						acctest.EqualField("scheme", o.Scheme, "https"),
						acctest.EqualField("host", o.Host, "eu-1.cdn77-storage.com"),
						acctest.NullFieldEqual("port", o.Port, 443),
						acctest.EqualField("bucket_name", o.BucketName, bucketName),
					)
				}),
			),
		},
		resource.TestStep{
			Config: acctest.Config(objectStoragesDataSourceConfig+`resource "cdn77_origin_object_storage" "os" {
					label = "{bucketName}"
					note = "some note"
					acl = "authenticated-read"
					cluster_id = local.eu_cluster_id
					bucket_name = "{bucketName}"
				}`, "bucketName", bucketName),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionDestroyBeforeCreate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAndReassignAttr(rsc, "id", &originId),
				resource.TestCheckResourceAttr(rsc, "label", bucketName),
				resource.TestCheckResourceAttr(rsc, "note", "some note"),
				resource.TestCheckResourceAttr(rsc, "url", "https://eu-1.cdn77-storage.com:443"),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", "https"),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", "eu-1.cdn77-storage.com"),
				resource.TestCheckResourceAttr(rsc, "url_parts.port", "443"),
				resource.TestCheckResourceAttr(rsc, "bucket_name", bucketName),
				resource.TestCheckResourceAttr(rsc, "acl", "authenticated-read"),
				acctest.CheckAndAssignAttr(rsc, "cluster_id", &clusterId),
				resource.TestCheckResourceAttr(rsc, "usage.files", "0"),
				resource.TestCheckResourceAttr(rsc, "usage.size_bytes", "0"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
				checkObjectStorage(client, &originId, func(o *cdn77.ObjectStorageOriginDetail) error {
					return errors.Join(
						acctest.EqualField("type", o.Type, origin.TypeObjectStorage),
						acctest.EqualField("label", o.Label, bucketName),
						acctest.NullFieldEqual("note", o.Note, "some note"),
						acctest.EqualField("scheme", o.Scheme, "https"),
						acctest.EqualField("host", o.Host, "eu-1.cdn77-storage.com"),
						acctest.NullFieldEqual("port", o.Port, 443),
						acctest.EqualField("bucket_name", o.BucketName, bucketName),
					)
				}),
			),
		},
		resource.TestStep{
			Config: acctest.Config(objectStoragesDataSourceConfig+`resource "cdn77_origin_object_storage" "os" {
					label = "{bucketName}"
					note = "some note"
					acl = "authenticated-read"
					cluster_id = local.us_cluster_id
					bucket_name = "{bucketName}"
				}`, "bucketName", bucketName),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionDestroyBeforeCreate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAndReassignAttr(rsc, "id", &originId),
				resource.TestCheckResourceAttr(rsc, "label", bucketName),
				resource.TestCheckResourceAttr(rsc, "note", "some note"),
				resource.TestCheckResourceAttr(rsc, "url", "https://us-1.cdn77-storage.com:443"),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", "https"),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", "us-1.cdn77-storage.com"),
				resource.TestCheckResourceAttr(rsc, "url_parts.port", "443"),
				resource.TestCheckResourceAttr(rsc, "bucket_name", bucketName),
				resource.TestCheckResourceAttr(rsc, "acl", "authenticated-read"),
				acctest.CheckAndReassignAttr(rsc, "cluster_id", &clusterId),
				resource.TestCheckResourceAttr(rsc, "usage.files", "0"),
				resource.TestCheckResourceAttr(rsc, "usage.size_bytes", "0"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
				checkObjectStorage(client, &originId, func(o *cdn77.ObjectStorageOriginDetail) error {
					return errors.Join(
						acctest.EqualField("type", o.Type, origin.TypeObjectStorage),
						acctest.EqualField("label", o.Label, bucketName),
						acctest.NullFieldEqual("note", o.Note, "some note"),
						acctest.EqualField("scheme", o.Scheme, "https"),
						acctest.EqualField("host", o.Host, "us-1.cdn77-storage.com"),
						acctest.NullFieldEqual("port", o.Port, 443),
						acctest.EqualField("bucket_name", o.BucketName, bucketName),
					)
				}),
			),
		},
		resource.TestStep{
			Config: acctest.Config(objectStoragesDataSourceConfig+`resource "cdn77_origin_object_storage" "os" {
					label = "{bucketName}"
					note = "some note"
					acl = "authenticated-read"
					cluster_id = local.us_cluster_id
					bucket_name = "{bucketName}"
				}`, "bucketName", anotherBucketName),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionDestroyBeforeCreate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAndReassignAttr(rsc, "id", &originId),
				resource.TestCheckResourceAttr(rsc, "label", anotherBucketName),
				resource.TestCheckResourceAttr(rsc, "note", "some note"),
				resource.TestCheckResourceAttr(rsc, "url", "https://us-1.cdn77-storage.com:443"),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", "https"),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", "us-1.cdn77-storage.com"),
				resource.TestCheckResourceAttr(rsc, "url_parts.port", "443"),
				resource.TestCheckResourceAttr(rsc, "bucket_name", anotherBucketName),
				resource.TestCheckResourceAttr(rsc, "acl", "authenticated-read"),
				acctest.CheckAttr(rsc, "cluster_id", &clusterId),
				resource.TestCheckResourceAttr(rsc, "usage.files", "0"),
				resource.TestCheckResourceAttr(rsc, "usage.size_bytes", "0"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
				checkObjectStorage(client, &originId, func(o *cdn77.ObjectStorageOriginDetail) error {
					return errors.Join(
						acctest.EqualField("type", o.Type, origin.TypeObjectStorage),
						acctest.EqualField("label", o.Label, anotherBucketName),
						acctest.NullFieldEqual("note", o.Note, "some note"),
						acctest.EqualField("scheme", o.Scheme, "https"),
						acctest.EqualField("host", o.Host, "us-1.cdn77-storage.com"),
						acctest.NullFieldEqual("port", o.Port, 443),
						acctest.EqualField("bucket_name", o.BucketName, anotherBucketName),
					)
				}),
			),
		},
	)
}

func TestAccOrigin_ObjectStorageResource_Import(t *testing.T) {
	const rsc = "cdn77_origin_object_storage.os"
	client := acctest.GetClient(t)
	bucketName := "my-bucket-" + uuid.New().String()
	var originId, clusterId string

	acctest.Run(t, acctest.CheckOriginDestroyed(client, origin.TypeObjectStorage),
		resource.TestStep{
			Config: acctest.Config(objectStoragesDataSourceConfig+`resource "cdn77_origin_object_storage" "os" {
					label = "{bucketName}"
					note = "some note"
					acl = "private"
					cluster_id = local.eu_cluster_id
					bucket_name = "{bucketName}"
				}`, "bucketName", bucketName),
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
	)
}

func TestAccOrigin_ObjectStorageDataSource_OnlyRequiredFields(t *testing.T) {
	const nonExistingOriginId = "bcd7b5bb-a044-4611-82e4-3f3b2a3cda13"
	const rsc = "data.cdn77_origin_object_storage.os"
	client := acctest.GetClient(t)
	originBucketName := "my-bucket-" + uuid.New().String()
	label := originBucketName
	request := cdn77.OriginCreateObjectStorageJSONRequestBody{
		Acl:        cdn77.AuthenticatedRead,
		BucketName: originBucketName,
		ClusterId:  "842b5641-b641-4723-ac81-f8cc286e288f",
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
		Acl:        cdn77.AuthenticatedRead,
		BucketName: originBucketName,
		ClusterId:  "842b5641-b641-4723-ac81-f8cc286e288f",
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

const objectStoragesDataSourceConfig = `
data "cdn77_object_storages" "all" {
}

locals {
	eu_cluster_id = one([for os in data.cdn77_object_storages.all.clusters : os.id if os.label == "EU"])
	us_cluster_id = one([for os in data.cdn77_object_storages.all.clusters : os.id if os.label == "US"])
}
`

const objectStorageDataSourceConfig = `
data "cdn77_origin_object_storage" "os" {
  id = "{id}"
}
`
