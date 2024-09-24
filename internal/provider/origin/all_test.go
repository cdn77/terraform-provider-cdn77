package origin_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/origin"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/shared"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oapi-codegen/nullable"
)

func TestAccOrigin_AllDataSource(t *testing.T) {
	const rsc = "data.cdn77_origins.all"
	client := acctest.GetClient(t)

	const urlLabel = "random origin"
	const urlNote = "some note"
	const urlScheme = "https"
	const urlHost = "my-totally-random-custom-host.com"
	const urlUrl = "https://my-totally-random-custom-host.com"
	urlRequest := cdn77.OriginCreateUrlJSONRequestBody{
		Label:  urlLabel,
		Note:   nullable.NewNullableWithValue(urlNote),
		Scheme: urlScheme,
		Host:   urlHost,
	}

	urlResponse, err := client.OriginCreateUrlWithResponse(context.Background(), urlRequest)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", urlResponse, err)

	urlId := urlResponse.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, origin.TypeUrl, urlId)
	})

	const awsLabel = "another origin"
	const awsScheme = "http"
	const awsHost = "some-other-totally-random-custom-host.com"
	const awsBasePath = "some-dir"
	const awsUrl = "http://some-other-totally-random-custom-host.com/some-dir"
	const awsAccessKeyId = "someKeyId"
	const awsAccessKeySecret = "someKeySecret"
	const awsRegion = "eu"
	awsRequest := cdn77.OriginCreateAwsJSONRequestBody{
		AwsAccessKeyId:     nullable.NewNullableWithValue(awsAccessKeyId),
		AwsAccessKeySecret: nullable.NewNullableWithValue(awsAccessKeySecret),
		AwsRegion:          nullable.NewNullableWithValue(awsRegion),
		BaseDir:            nullable.NewNullableWithValue(awsBasePath),
		Host:               awsHost,
		Label:              awsLabel,
		Scheme:             awsScheme,
	}

	awsResponse, err := client.OriginCreateAwsWithResponse(context.Background(), awsRequest)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", awsResponse, err)

	awsId := awsResponse.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, origin.TypeAws, awsId)
	})

	const osLabel = "yet another origin"
	const osNote = "just a note"
	osBucketName := "my-bucket-" + uuid.New().String()
	osRequest := cdn77.OriginCreateObjectStorageJSONRequestBody{
		Acl:        cdn77.AuthenticatedRead,
		BucketName: osBucketName,
		ClusterId:  "842b5641-b641-4723-ac81-f8cc286e288f",
		Label:      osLabel,
		Note:       nullable.NewNullableWithValue(osNote),
	}

	osResponse, err := client.OriginCreateObjectStorageWithResponse(context.Background(), osRequest)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", osResponse, err)

	osId := osResponse.JSON201.Id
	osScheme := string(osResponse.JSON201.Scheme)
	osHost := osResponse.JSON201.Host
	osPort := osResponse.JSON201.Port
	osUrlModel := shared.NewUrlModel(context.Background(), osScheme, osHost, osPort, nullable.NewNullNullable[string]())
	osUrl := osUrlModel.Url.ValueString()

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, origin.TypeObjectStorage, osId)
	})

	acctest.Run(t, nil, resource.TestStep{
		Config: allDataSourceConfig,
		Check: resource.ComposeAggregateTestCheckFunc(
			// URL origin
			resource.TestCheckResourceAttr(rsc, "url.#", "1"),
			resource.TestCheckResourceAttr(rsc, "url.0.id", urlId),
			resource.TestCheckResourceAttr(rsc, "url.0.label", urlLabel),
			resource.TestCheckResourceAttr(rsc, "url.0.note", urlNote),
			resource.TestCheckResourceAttr(rsc, "url.0.url", urlUrl),
			resource.TestCheckResourceAttr(rsc, "url.0.url_parts.scheme", urlScheme),
			resource.TestCheckResourceAttr(rsc, "url.0.url_parts.host", urlHost),
			resource.TestCheckNoResourceAttr(rsc, "url.0.url_parts.port"),
			resource.TestCheckNoResourceAttr(rsc, "url.0.url_parts.base_path"),

			// AWS origin
			resource.TestCheckResourceAttr(rsc, "aws.#", "1"),
			resource.TestCheckResourceAttr(rsc, "aws.0.id", awsId),
			resource.TestCheckResourceAttr(rsc, "aws.0.label", awsLabel),
			resource.TestCheckResourceAttr(rsc, "aws.0.url", awsUrl),
			resource.TestCheckResourceAttr(rsc, "aws.0.url_parts.scheme", awsScheme),
			resource.TestCheckResourceAttr(rsc, "aws.0.url_parts.host", awsHost),
			resource.TestCheckResourceAttr(rsc, "aws.0.url_parts.base_path", awsBasePath),
			resource.TestCheckResourceAttr(rsc, "aws.0.access_key_id", awsAccessKeyId),
			resource.TestCheckResourceAttr(rsc, "aws.0.region", awsRegion),
			resource.TestCheckNoResourceAttr(rsc, "aws.0.note"),
			resource.TestCheckNoResourceAttr(rsc, "aws.0.url_parts.port"),
			resource.TestCheckNoResourceAttr(rsc, "aws.0.access_key_secret"),

			// Object Storage Origin
			resource.TestCheckResourceAttr(rsc, "object_storage.#", "1"),
			resource.TestCheckResourceAttr(rsc, "object_storage.0.id", osId),
			resource.TestCheckResourceAttr(rsc, "object_storage.0.label", osLabel),
			resource.TestCheckResourceAttr(rsc, "object_storage.0.note", osNote),
			resource.TestCheckResourceAttr(rsc, "object_storage.0.url", osUrl),
			resource.TestCheckResourceAttr(rsc, "object_storage.0.url_parts.scheme", osScheme),
			resource.TestCheckResourceAttr(rsc, "object_storage.0.url_parts.host", osHost),
			func(state *terraform.State) error {
				if port, err := osPort.Get(); err == nil {
					return resource.TestCheckResourceAttr(
						rsc, "object_storage.0.url_parts.port", strconv.Itoa(port),
					)(state)
				}

				return resource.TestCheckNoResourceAttr(rsc, "object_storage.0.url_parts.port")(state)
			},
			resource.TestCheckResourceAttr(rsc, "object_storage.0.bucket_name", osBucketName),
			resource.TestCheckNoResourceAttr(rsc, "object_storage.0.url_parts.base_path"),
			resource.TestCheckNoResourceAttr(rsc, "object_storage.0.acl"),
			resource.TestCheckNoResourceAttr(rsc, "object_storage.0.cluster_id"),
			resource.TestCheckNoResourceAttr(rsc, "object_storage.0.access_key_id"),
			resource.TestCheckNoResourceAttr(rsc, "object_storage.0.access_key_secret"),
		),
	})
}

func TestAccOrigin_AllDataSource_Empty(t *testing.T) {
	acctest.Run(t, nil, resource.TestStep{
		Config: allDataSourceConfig,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.cdn77_origins.all", "aws.#", "0"),
			resource.TestCheckResourceAttr("data.cdn77_origins.all", "object_storage.#", "0"),
			resource.TestCheckResourceAttr("data.cdn77_origins.all", "url.#", "0"),
		),
	})
}

const allDataSourceConfig = `
data "cdn77_origins" "all" {
}
`
