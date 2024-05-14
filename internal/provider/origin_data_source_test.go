package provider_test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oapi-codegen/nullable"
)

func TestAccOriginDataSource_NonExistingOrigin(t *testing.T) {
	const originId = "bcd7b5bb-a044-4611-82e4-3f3b2a3cda13"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      acctest.Config(originDataSourceConfigAws, "id", originId),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`.*?"%s".*?not found.*?`, originId)),
			},
			{
				Config:      acctest.Config(originDataSourceConfigObjectStorage, "id", originId),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`.*?"%s".*?not found.*?`, originId)),
			},
			{
				Config:      acctest.Config(originDataSourceConfigUrl, "id", originId),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`.*?"%s".*?not found.*?`, originId)),
			},
		},
	})
}

func TestAccOriginDataSource_Aws_OnlyRequiredFields(t *testing.T) {
	client := acctest.GetClient(t)

	const rsc = "data.cdn77_origin.aws"
	const originHost = "my-totally-random-custom-host.com"
	const originLabel = "random origin"
	const originScheme = "https"

	request := cdn77.OriginAddAwsJSONRequestBody{
		Host:   originHost,
		Label:  originLabel,
		Scheme: originScheme,
	}
	response, err := client.OriginAddAwsWithResponse(context.Background(), request)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response, err)

	originId := response.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, provider.OriginTypeAws, originId)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: acctest.Config(originDataSourceConfigAws, "id", originId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rsc, "id", originId),
					resource.TestCheckResourceAttr(rsc, "type", provider.OriginTypeAws),
					resource.TestCheckResourceAttr(rsc, "label", originLabel),
					resource.TestCheckResourceAttr(rsc, "scheme", originScheme),
					resource.TestCheckResourceAttr(rsc, "host", originHost),
					resource.TestCheckNoResourceAttr(rsc, "note"),
					resource.TestCheckNoResourceAttr(rsc, "aws_access_key_id"),
					resource.TestCheckNoResourceAttr(rsc, "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr(rsc, "aws_region"),
					resource.TestCheckNoResourceAttr(rsc, "acl"),
					resource.TestCheckNoResourceAttr(rsc, "cluster_id"),
					resource.TestCheckNoResourceAttr(rsc, "bucket_name"),
					resource.TestCheckNoResourceAttr(rsc, "access_key_id"),
					resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
					resource.TestCheckNoResourceAttr(rsc, "port"),
					resource.TestCheckNoResourceAttr(rsc, "base_dir"),
				),
			},
		},
	})
}

func TestAccOriginDataSource_Aws_AllFields(t *testing.T) {
	client := acctest.GetClient(t)

	const rsc = "data.cdn77_origin.aws"
	const originAwsKeyId = "someKeyId"
	const originAwsKeySecret = "someKeySecret"
	const originAwsRegion = "eu"
	const originBaseDir = "some-dir"
	const originHost = "my-totally-random-custom-host.com"
	const originLabel = "random origin"
	const originNote = "some note"
	const originPort = 12345
	const originScheme = "https"

	request := cdn77.OriginAddAwsJSONRequestBody{
		AwsAccessKeyId:     nullable.NewNullableWithValue(originAwsKeyId),
		AwsAccessKeySecret: nullable.NewNullableWithValue(originAwsKeySecret),
		AwsRegion:          nullable.NewNullableWithValue(originAwsRegion),
		BaseDir:            nullable.NewNullableWithValue(originBaseDir),
		Host:               originHost,
		Label:              originLabel,
		Note:               nullable.NewNullableWithValue(originNote),
		Port:               nullable.NewNullableWithValue(originPort),
		Scheme:             originScheme,
	}
	response, err := client.OriginAddAwsWithResponse(context.Background(), request)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response, err)

	originId := response.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, provider.OriginTypeAws, originId)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: acctest.Config(originDataSourceConfigAws, "id", originId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rsc, "id", originId),
					resource.TestCheckResourceAttr(rsc, "type", provider.OriginTypeAws),
					resource.TestCheckResourceAttr(rsc, "label", originLabel),
					resource.TestCheckResourceAttr(rsc, "note", originNote),
					resource.TestCheckResourceAttr(rsc, "aws_access_key_id", originAwsKeyId),
					resource.TestCheckResourceAttr(rsc, "aws_region", originAwsRegion),
					resource.TestCheckResourceAttr(rsc, "host", originHost),
					resource.TestCheckResourceAttr(rsc, "port", strconv.Itoa(originPort)),
					resource.TestCheckResourceAttr(rsc, "base_dir", originBaseDir),
					resource.TestCheckNoResourceAttr(rsc, "acl"),
					resource.TestCheckNoResourceAttr(rsc, "cluster_id"),
					resource.TestCheckNoResourceAttr(rsc, "bucket_name"),
					resource.TestCheckNoResourceAttr(rsc, "access_key_id"),
					resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
				),
			},
		},
	})
}

func TestAccOriginDataSource_ObjectStorage_OnlyRequiredFields(t *testing.T) {
	client := acctest.GetClient(t)

	const rsc = "data.cdn77_origin.os"
	originBucketName := "my-bucket-" + uuid.New().String()
	const originLabel = "random origin"

	request := cdn77.OriginAddObjectStorageJSONRequestBody{
		Acl:        cdn77.AuthenticatedRead,
		BucketName: originBucketName,
		ClusterId:  "842b5641-b641-4723-ac81-f8cc286e288f",
		Label:      originLabel,
	}
	response, err := client.OriginAddObjectStorageWithResponse(context.Background(), request)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response, err)

	originId := response.JSON201.Id
	originScheme := string(response.JSON201.Scheme)
	originHost := response.JSON201.Host
	originPort := response.JSON201.Port

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, provider.OriginTypeObjectStorage, originId)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: acctest.Config(originDataSourceConfigObjectStorage, "id", originId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rsc, "id", originId),
					resource.TestCheckResourceAttr(rsc, "type", provider.OriginTypeObjectStorage),
					resource.TestCheckResourceAttr(rsc, "label", originLabel),
					resource.TestCheckResourceAttr(rsc, "bucket_name", originBucketName),
					resource.TestCheckResourceAttr(rsc, "scheme", originScheme),
					resource.TestCheckResourceAttr(rsc, "host", originHost),
					func(state *terraform.State) error {
						if originPort.IsNull() {
							return resource.TestCheckNoResourceAttr(rsc, "port")(state)
						}

						return resource.TestCheckResourceAttr(rsc, "port", strconv.Itoa(originPort.MustGet()))(state)
					},
					resource.TestCheckNoResourceAttr(rsc, "note"),
					resource.TestCheckNoResourceAttr(rsc, "aws_access_key_id"),
					resource.TestCheckNoResourceAttr(rsc, "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr(rsc, "aws_region"),
					resource.TestCheckNoResourceAttr(rsc, "acl"),
					resource.TestCheckNoResourceAttr(rsc, "cluster_id"),
					resource.TestCheckNoResourceAttr(rsc, "access_key_id"),
					resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
					resource.TestCheckNoResourceAttr(rsc, "base_dir"),
				),
			},
		},
	})
}

func TestAccOriginDataSource_ObjectStorage_AllFields(t *testing.T) {
	client := acctest.GetClient(t)

	const rsc = "data.cdn77_origin.os"
	originBucketName := "my-bucket-" + uuid.New().String()
	const originLabel = "random origin"
	const originNote = "some note"

	request := cdn77.OriginAddObjectStorageJSONRequestBody{
		Acl:        cdn77.AuthenticatedRead,
		BucketName: originBucketName,
		ClusterId:  "842b5641-b641-4723-ac81-f8cc286e288f",
		Label:      originLabel,
		Note:       nullable.NewNullableWithValue(originNote),
	}

	response, err := client.OriginAddObjectStorageWithResponse(context.Background(), request)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response, err)

	originId := response.JSON201.Id
	originScheme := string(response.JSON201.Scheme)
	originHost := response.JSON201.Host
	originPort := response.JSON201.Port

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, provider.OriginTypeObjectStorage, originId)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: acctest.Config(originDataSourceConfigObjectStorage, "id", originId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rsc, "id", originId),
					resource.TestCheckResourceAttr(rsc, "type", provider.OriginTypeObjectStorage),
					resource.TestCheckResourceAttr(rsc, "label", originLabel),
					resource.TestCheckResourceAttr(rsc, "note", originNote),
					resource.TestCheckResourceAttr(rsc, "bucket_name", originBucketName),
					resource.TestCheckResourceAttr(rsc, "scheme", originScheme),
					resource.TestCheckResourceAttr(rsc, "host", originHost),
					func(state *terraform.State) error {
						if originPort.IsNull() {
							return resource.TestCheckNoResourceAttr(rsc, "port")(state)
						}

						return resource.TestCheckResourceAttr(rsc, "port", strconv.Itoa(originPort.MustGet()))(state)
					},
					resource.TestCheckNoResourceAttr(rsc, "aws_access_key_id"),
					resource.TestCheckNoResourceAttr(rsc, "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr(rsc, "aws_region"),
					resource.TestCheckNoResourceAttr(rsc, "acl"),
					resource.TestCheckNoResourceAttr(rsc, "cluster_id"),
					resource.TestCheckNoResourceAttr(rsc, "access_key_id"),
					resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
					resource.TestCheckNoResourceAttr(rsc, "base_dir"),
				),
			},
		},
	})
}

func TestAccOriginDataSource_Url_OnlyRequiredFields(t *testing.T) {
	client := acctest.GetClient(t)

	const rsc = "data.cdn77_origin.url"
	const originHost = "my-totally-random-custom-host.com"
	const originLabel = "random origin"
	const originScheme = "https"

	request := cdn77.OriginAddUrlJSONRequestBody{
		Host:   originHost,
		Label:  originLabel,
		Scheme: originScheme,
	}
	response, err := client.OriginAddUrlWithResponse(context.Background(), request)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response, err)

	originId := response.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, provider.OriginTypeUrl, originId)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: acctest.Config(originDataSourceConfigUrl, "id", originId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rsc, "id", originId),
					resource.TestCheckResourceAttr(rsc, "type", provider.OriginTypeUrl),
					resource.TestCheckResourceAttr(rsc, "label", originLabel),
					resource.TestCheckResourceAttr(rsc, "scheme", originScheme),
					resource.TestCheckResourceAttr(rsc, "host", originHost),
					resource.TestCheckNoResourceAttr(rsc, "note"),
					resource.TestCheckNoResourceAttr(rsc, "aws_access_key_id"),
					resource.TestCheckNoResourceAttr(rsc, "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr(rsc, "aws_region"),
					resource.TestCheckNoResourceAttr(rsc, "acl"),
					resource.TestCheckNoResourceAttr(rsc, "cluster_id"),
					resource.TestCheckNoResourceAttr(rsc, "bucket_name"),
					resource.TestCheckNoResourceAttr(rsc, "access_key_id"),
					resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
					resource.TestCheckNoResourceAttr(rsc, "port"),
					resource.TestCheckNoResourceAttr(rsc, "base_dir"),
				),
			},
		},
	})
}

func TestAccOriginDataSource_Url_AllFields(t *testing.T) {
	client := acctest.GetClient(t)

	const rsc = "data.cdn77_origin.url"
	const originBaseDir = "some-dir"
	const originHost = "my-totally-random-custom-host.com"
	const originLabel = "random origin"
	const originNote = "some note"
	const originPort = 12345
	const originScheme = "https"

	request := cdn77.OriginAddUrlJSONRequestBody{
		BaseDir: nullable.NewNullableWithValue(originBaseDir),
		Host:    originHost,
		Label:   originLabel,
		Note:    nullable.NewNullableWithValue(originNote),
		Port:    nullable.NewNullableWithValue(originPort),
		Scheme:  originScheme,
	}
	response, err := client.OriginAddUrlWithResponse(context.Background(), request)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response, err)

	originId := response.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, provider.OriginTypeUrl, originId)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: acctest.Config(originDataSourceConfigUrl, "id", originId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rsc, "id", originId),
					resource.TestCheckResourceAttr(rsc, "type", provider.OriginTypeUrl),
					resource.TestCheckResourceAttr(rsc, "label", originLabel),
					resource.TestCheckResourceAttr(rsc, "note", originNote),
					resource.TestCheckResourceAttr(rsc, "host", originHost),
					resource.TestCheckResourceAttr(rsc, "port", strconv.Itoa(originPort)),
					resource.TestCheckResourceAttr(rsc, "base_dir", originBaseDir),
					resource.TestCheckNoResourceAttr(rsc, "aws_access_key_id"),
					resource.TestCheckNoResourceAttr(rsc, "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr(rsc, "aws_region"),
					resource.TestCheckNoResourceAttr(rsc, "acl"),
					resource.TestCheckNoResourceAttr(rsc, "cluster_id"),
					resource.TestCheckNoResourceAttr(rsc, "bucket_name"),
					resource.TestCheckNoResourceAttr(rsc, "access_key_id"),
					resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
				),
			},
		},
	})
}

const originDataSourceConfigAws = `
data "cdn77_origin" "aws" {
  id = "{id}"
  type = "aws"
}
`

const originDataSourceConfigObjectStorage = `
data "cdn77_origin" "os" {
  id = "{id}"
  type = "object-storage"
}
`

const originDataSourceConfigUrl = `
data "cdn77_origin" "url" {
  id = "{id}"
  type = "url"
}
`
