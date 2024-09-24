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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oapi-codegen/nullable"
)

func TestAccOrigin_AwsResource(t *testing.T) {
	const rsc = "cdn77_origin_aws.aws"
	client := acctest.GetClient(t)
	var originId string

	acctest.Run(t, acctest.CheckOriginDestroyed(client, origin.TypeAws),
		resource.TestStep{
			Config: `resource "cdn77_origin_aws" "aws" {
				label = "some label"
				url = "http://my-totally-random-custom-host.com"
				url_parts = {
					scheme = "http"
					host = "my-totally-random-custom-host.com"
				}
			}`,
			ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Combination.*` +
				`attributes specified when one \(and only one\) of \[url,url_parts\] is required`),
			PlanOnly: true,
		},
		resource.TestStep{
			Config: `resource "cdn77_origin_aws" "aws" {
				label = "some label"
				url = "http://my-totally-random-custom-host.com"
			}`,
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionCreate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAndAssignAttr(rsc, "id", &originId),
				resource.TestCheckResourceAttr(rsc, "label", "some label"),
				resource.TestCheckResourceAttr(rsc, "url", "http://my-totally-random-custom-host.com"),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", "http"),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", "my-totally-random-custom-host.com"),
				resource.TestCheckNoResourceAttr(rsc, "note"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.port"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
				resource.TestCheckNoResourceAttr(rsc, "access_key_id"),
				resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
				resource.TestCheckNoResourceAttr(rsc, "region"),
				checkAws(client, &originId, func(o *cdn77.S3OriginDetail) error {
					return errors.Join(
						acctest.EqualField("type", o.Type, origin.TypeAws),
						acctest.EqualField("label", o.Label, "some label"),
						acctest.NullField("note", o.Note),
						acctest.EqualField("scheme", o.Scheme, "http"),
						acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
						acctest.NullField("port", o.Port),
						acctest.NullField("base_dir", o.BaseDir),
						acctest.NullField("access_key_id", o.AwsAccessKeyId),
						acctest.NullField("region", o.AwsRegion),
					)
				}),
			),
		},
		resource.TestStep{
			Config: `resource "cdn77_origin_aws" "aws" {
				label = "some label"
				url_parts = {
					scheme = "http"
					host = "my-totally-random-custom-host.com"
				}
			}`,
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionNoop),
		},
		resource.TestStep{
			Config: `resource "cdn77_origin_aws" "aws" {
				label = "another label"
				note = "some note"
				url = "http://my-totally-random-custom-host.com:12345"
			}`,
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionUpdate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAttr(rsc, "id", &originId),
				resource.TestCheckResourceAttr(rsc, "label", "another label"),
				resource.TestCheckResourceAttr(rsc, "note", "some note"),
				resource.TestCheckResourceAttr(rsc, "url", "http://my-totally-random-custom-host.com:12345"),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", "http"),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", "my-totally-random-custom-host.com"),
				resource.TestCheckResourceAttr(rsc, "url_parts.port", "12345"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
				resource.TestCheckNoResourceAttr(rsc, "access_key_id"),
				resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
				resource.TestCheckNoResourceAttr(rsc, "region"),
				checkAws(client, &originId, func(o *cdn77.S3OriginDetail) error {
					return errors.Join(
						acctest.EqualField("type", o.Type, origin.TypeAws),
						acctest.EqualField("label", o.Label, "another label"),
						acctest.NullFieldEqual("note", o.Note, "some note"),
						acctest.EqualField("scheme", o.Scheme, "http"),
						acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
						acctest.NullFieldEqual("port", o.Port, 12345),
						acctest.NullField("base_dir", o.BaseDir),
						acctest.NullField("access_key_id", o.AwsAccessKeyId),
						acctest.NullField("region", o.AwsRegion),
					)
				}),
			),
		},
		resource.TestStep{
			Config: `resource "cdn77_origin_aws" "aws" {
				label = "another label"
				note = "some note"
				url = "http://my-totally-random-custom-host.com:12345/some-dir"
				access_key_id = "keyid"
				access_key_secret = "keysecret"
				region = "eu"
			}`,
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionUpdate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAttr(rsc, "id", &originId),
				resource.TestCheckResourceAttr(rsc, "label", "another label"),
				resource.TestCheckResourceAttr(rsc, "note", "some note"),
				resource.TestCheckResourceAttr(rsc, "access_key_id", "keyid"),
				resource.TestCheckResourceAttr(rsc, "access_key_secret", "keysecret"),
				resource.TestCheckResourceAttr(rsc, "region", "eu"),
				resource.TestCheckResourceAttr(rsc, "url", "http://my-totally-random-custom-host.com:12345/some-dir"),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", "http"),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", "my-totally-random-custom-host.com"),
				resource.TestCheckResourceAttr(rsc, "url_parts.port", "12345"),
				resource.TestCheckResourceAttr(rsc, "url_parts.base_path", "some-dir"),
				checkAws(client, &originId, func(o *cdn77.S3OriginDetail) error {
					return errors.Join(
						acctest.EqualField("type", o.Type, origin.TypeAws),
						acctest.EqualField("label", o.Label, "another label"),
						acctest.NullFieldEqual("note", o.Note, "some note"),
						acctest.EqualField("scheme", o.Scheme, "http"),
						acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
						acctest.NullFieldEqual("port", o.Port, 12345),
						acctest.NullFieldEqual("base_dir", o.BaseDir, "some-dir"),
						acctest.NullFieldEqual("access_key_id", o.AwsAccessKeyId, "keyid"),
						acctest.NullFieldEqual("region", o.AwsRegion, "eu"),
					)
				}),
			),
		},
		resource.TestStep{
			Config: `resource "cdn77_origin_aws" "aws" {
				label = "another label"
				note = "some note"
				url_parts = {
					scheme = "http"
					host = "my-totally-random-custom-host.com"
					port = 12345
					base_path = "some-dir"
				}
				access_key_id = "keyid"
				access_key_secret = "keysecret"
				region = "eu"
			}`,
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionNoop),
		},
		resource.TestStep{
			Config: `resource "cdn77_origin_aws" "aws" {
				label = "another label"
				url = "http://my-totally-random-custom-host.com"
			}`,
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionUpdate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAttr(rsc, "id", &originId),
				resource.TestCheckResourceAttr(rsc, "label", "another label"),
				resource.TestCheckResourceAttr(rsc, "url", "http://my-totally-random-custom-host.com"),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", "http"),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", "my-totally-random-custom-host.com"),
				resource.TestCheckNoResourceAttr(rsc, "note"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.port"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
				resource.TestCheckNoResourceAttr(rsc, "access_key_id"),
				resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
				resource.TestCheckNoResourceAttr(rsc, "region"),
				checkAws(client, &originId, func(o *cdn77.S3OriginDetail) error {
					return errors.Join(
						acctest.EqualField("type", o.Type, origin.TypeAws),
						acctest.EqualField("label", o.Label, "another label"),
						acctest.NullField("note", o.Note),
						acctest.EqualField("scheme", o.Scheme, "http"),
						acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
						acctest.NullField("port", o.Port),
						acctest.NullField("base_dir", o.BaseDir),
						acctest.NullField("access_key_id", o.AwsAccessKeyId),
						acctest.NullField("region", o.AwsRegion),
					)
				}),
			),
		},
	)
}

func TestAccOrigin_AwsResource_Import(t *testing.T) {
	const rsc = "cdn77_origin_aws.aws"
	client := acctest.GetClient(t)
	var originId string

	acctest.Run(t, acctest.CheckOriginDestroyed(client, origin.TypeAws),
		resource.TestStep{
			Config: `resource "cdn77_origin_aws" "aws" {
					label = "some label"
					note = "some note"
					url = "http://my-totally-random-custom-host.com"
					access_key_id = "keyid"
					access_key_secret = "keysecret"
					region = "eu"
				}`,
			Check: acctest.CheckAndAssignAttr(rsc, "id", &originId),
		},
		resource.TestStep{
			ResourceName: rsc,
			ImportState:  true,
			ImportStateIdFunc: func(*terraform.State) (string, error) {
				return fmt.Sprintf("%s,keysecret", originId), nil
			},
			ImportStateVerify: true,
		},
	)
}

func TestAccOrigin_AwsResource_Import_NoKey(t *testing.T) {
	const rsc = "cdn77_origin_aws.aws"
	client := acctest.GetClient(t)
	var originId string

	acctest.Run(t, acctest.CheckOriginDestroyed(client, origin.TypeAws),
		resource.TestStep{
			Config: `resource "cdn77_origin_aws" "aws" {
					label = "some label"
					note = "some note"
					url = "http://my-totally-random-custom-host.com"
				}`,
			Check: acctest.CheckAndAssignAttr(rsc, "id", &originId),
		},
		resource.TestStep{
			ResourceName: rsc,
			ImportState:  true,
			ImportStateIdFunc: func(*terraform.State) (string, error) {
				return fmt.Sprintf("%s,", originId), nil
			},
			ImportStateVerify: true,
		},
	)
}

func TestAccOrigin_AwsDataSource_OnlyRequiredFields(t *testing.T) {
	const nonExistingOriginId = "bcd7b5bb-a044-4611-82e4-3f3b2a3cda13"
	const rsc = "data.cdn77_origin_aws.aws"
	const label = "random origin"
	const originUrl = "http://my-totally-random-custom-host.com"
	const scheme = "http"
	const host = "my-totally-random-custom-host.com"
	client := acctest.GetClient(t)
	request := cdn77.OriginCreateAwsJSONRequestBody{
		Label:  label,
		Scheme: scheme,
		Host:   host,
	}

	response, err := client.OriginCreateAwsWithResponse(context.Background(), request)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response, err)

	originId := response.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, origin.TypeAws, originId)
	})

	acctest.Run(t, nil,
		resource.TestStep{
			Config:      acctest.Config(awsDataSourceConfig, "id", nonExistingOriginId),
			ExpectError: regexp.MustCompile(fmt.Sprintf(`.*?"%s".*?not found.*?`, nonExistingOriginId)),
		},
		resource.TestStep{
			Config: acctest.Config(awsDataSourceConfig, "id", originId),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(rsc, "id", originId),
				resource.TestCheckResourceAttr(rsc, "label", label),
				resource.TestCheckResourceAttr(rsc, "url", originUrl),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", scheme),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", host),
				resource.TestCheckNoResourceAttr(rsc, "note"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.port"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
				resource.TestCheckNoResourceAttr(rsc, "access_key_id"),
				resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
				resource.TestCheckNoResourceAttr(rsc, "region"),
			),
		},
	)
}

func TestAccOrigin_AwsDataSource_AllFields(t *testing.T) {
	const rsc = "data.cdn77_origin_aws.aws"
	const label = "random origin"
	const note = "some note"
	const originUrl = "https://my-totally-random-custom-host.com:12345/some-dir"
	const scheme = "https"
	const host = "my-totally-random-custom-host.com"
	const port = 12345
	const basePath = "some-dir"
	const accessKeyId = "someKeyId"
	const awsKeySecret = "someKeySecret"
	const region = "eu"
	client := acctest.GetClient(t)
	request := cdn77.OriginCreateAwsJSONRequestBody{
		Label:              label,
		Note:               nullable.NewNullableWithValue(note),
		Scheme:             scheme,
		Host:               host,
		Port:               nullable.NewNullableWithValue(port),
		BaseDir:            nullable.NewNullableWithValue(basePath),
		AwsAccessKeyId:     nullable.NewNullableWithValue(accessKeyId),
		AwsAccessKeySecret: nullable.NewNullableWithValue(awsKeySecret),
		AwsRegion:          nullable.NewNullableWithValue(region),
	}

	response, err := client.OriginCreateAwsWithResponse(context.Background(), request)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response, err)

	originId := response.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, origin.TypeAws, originId)
	})

	acctest.Run(t, nil, resource.TestStep{
		Config: acctest.Config(awsDataSourceConfig, "id", originId),
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(rsc, "id", originId),
			resource.TestCheckResourceAttr(rsc, "label", label),
			resource.TestCheckResourceAttr(rsc, "note", note),
			resource.TestCheckResourceAttr(rsc, "url", originUrl),
			resource.TestCheckResourceAttr(rsc, "url_parts.scheme", scheme),
			resource.TestCheckResourceAttr(rsc, "url_parts.host", host),
			resource.TestCheckResourceAttr(rsc, "url_parts.port", strconv.Itoa(port)),
			resource.TestCheckResourceAttr(rsc, "url_parts.base_path", basePath),
			resource.TestCheckResourceAttr(rsc, "access_key_id", accessKeyId),
			resource.TestCheckResourceAttr(rsc, "region", region),
			resource.TestCheckNoResourceAttr(rsc, "access_key_secret"),
		),
	})
}

func checkAws(
	client cdn77.ClientWithResponsesInterface,
	originId *string,
	fn func(o *cdn77.S3OriginDetail) error,
) func(*terraform.State) error {
	return func(*terraform.State) error {
		response, err := client.OriginDetailAwsWithResponse(context.Background(), *originId)
		message := fmt.Sprintf("failed to get Origin[id=%s]: %%s", *originId)

		if err = acctest.CheckResponse(message, response, err); err != nil {
			return err
		}

		return fn(response.JSON200)
	}
}

const awsDataSourceConfig = `
data "cdn77_origin_aws" "aws" {
  id = "{id}"
}
`
