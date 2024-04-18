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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/oapi-codegen/nullable"
)

func TestAccOriginDataSource_NonExistingOrigin(t *testing.T) {
	const originId = "bcd7b5bb-a044-4611-82e4-3f3b2a3cda13"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      acctest.Config(originDataSourceConfigAws, "id", originId),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`Storage with id:\W+"%s" could not be found`, originId)),
			},
		},
	})
}

func TestAccOriginDataSource_Aws_OnlyRequiredFields(t *testing.T) {
	client := acctest.GetClient(t)

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
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "id", originId),
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "type", provider.OriginTypeAws),
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "label", originLabel),
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "scheme", originScheme),
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "host", originHost),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.aws", "note"),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.aws", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.aws", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.aws", "aws_region"),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.aws", "port"),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.aws", "base_dir"),
				),
			},
		},
	})
}

func TestAccOriginDataSource_Aws_AllFields(t *testing.T) {
	client := acctest.GetClient(t)

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
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "id", originId),
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "type", provider.OriginTypeAws),
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "label", originLabel),
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "note", originNote),
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "aws_access_key_id", originAwsKeyId),
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "aws_region", originAwsRegion),
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "host", originHost),
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "port", strconv.Itoa(originPort)),
					resource.TestCheckResourceAttr("data.cdn77_origin.aws", "base_dir", originBaseDir),
				),
			},
		},
	})
}

func TestAccOriginDataSource_Url_OnlyRequiredFields(t *testing.T) {
	client := acctest.GetClient(t)

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
					resource.TestCheckResourceAttr("data.cdn77_origin.url", "id", originId),
					resource.TestCheckResourceAttr("data.cdn77_origin.url", "type", provider.OriginTypeUrl),
					resource.TestCheckResourceAttr("data.cdn77_origin.url", "label", originLabel),
					resource.TestCheckResourceAttr("data.cdn77_origin.url", "scheme", originScheme),
					resource.TestCheckResourceAttr("data.cdn77_origin.url", "host", originHost),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.url", "note"),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.url", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.url", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.url", "aws_region"),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.url", "port"),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.url", "base_dir"),
				),
			},
		},
	})
}

func TestAccOriginDataSource_Url_AllFields(t *testing.T) {
	client := acctest.GetClient(t)

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
					resource.TestCheckResourceAttr("data.cdn77_origin.url", "id", originId),
					resource.TestCheckResourceAttr("data.cdn77_origin.url", "type", provider.OriginTypeUrl),
					resource.TestCheckResourceAttr("data.cdn77_origin.url", "label", originLabel),
					resource.TestCheckResourceAttr("data.cdn77_origin.url", "note", originNote),
					resource.TestCheckResourceAttr("data.cdn77_origin.url", "host", originHost),
					resource.TestCheckResourceAttr("data.cdn77_origin.url", "port", strconv.Itoa(originPort)),
					resource.TestCheckResourceAttr("data.cdn77_origin.url", "base_dir", originBaseDir),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.url", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.url", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("data.cdn77_origin.url", "aws_region"),
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

const originDataSourceConfigUrl = `
data "cdn77_origin" "url" {
  id = "{id}"
  type = "url"
}
`
