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

func TestAccOrigin_UrlResource(t *testing.T) {
	client := acctest.GetClient(t)
	rsc := "cdn77_origin_url.url"
	var originId string

	acctest.Run(t, acctest.CheckOriginDestroyed(client, origin.TypeUrl),
		resource.TestStep{
			Config: `resource "cdn77_origin_url" "url" {
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
			Config: `resource "cdn77_origin_url" "url" {
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
				checkUrl(client, &originId, func(o *cdn77.UrlOriginDetail) error {
					return errors.Join(
						acctest.EqualField("type", o.Type, origin.TypeUrl),
						acctest.EqualField("label", o.Label, "some label"),
						acctest.NullField("note", o.Note),
						acctest.EqualField("scheme", o.Scheme, "http"),
						acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
						acctest.NullField("port", o.Port),
						acctest.NullField("base_dir", o.BaseDir),
					)
				}),
			),
		},
		resource.TestStep{
			Config: `resource "cdn77_origin_url" "url" {
				label = "some label"
				url_parts = {
					scheme = "http"
					host = "my-totally-random-custom-host.com"
				}
			}`,
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionNoop),
		},
		resource.TestStep{
			Config: `resource "cdn77_origin_url" "url" {
				label = "another label"
				note = "some note"
				url = "http://my-totally-random-custom-host.com:12345/some-dir"
			}`,
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionUpdate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAttr(rsc, "id", &originId),
				resource.TestCheckResourceAttr(rsc, "label", "another label"),
				resource.TestCheckResourceAttr(rsc, "note", "some note"),
				resource.TestCheckResourceAttr(rsc, "url", "http://my-totally-random-custom-host.com:12345/some-dir"),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", "http"),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", "my-totally-random-custom-host.com"),
				resource.TestCheckResourceAttr(rsc, "url_parts.port", "12345"),
				resource.TestCheckResourceAttr(rsc, "url_parts.base_path", "some-dir"),
				checkUrl(client, &originId, func(o *cdn77.UrlOriginDetail) error {
					return errors.Join(
						acctest.EqualField("type", o.Type, origin.TypeUrl),
						acctest.EqualField("label", o.Label, "another label"),
						acctest.NullFieldEqual("note", o.Note, "some note"),
						acctest.EqualField("scheme", o.Scheme, "http"),
						acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
						acctest.NullFieldEqual("port", o.Port, 12345),
						acctest.NullFieldEqual("base_dir", o.BaseDir, "some-dir"),
					)
				}),
			),
		},
		resource.TestStep{
			Config: `resource "cdn77_origin_url" "url" {
				label = "another label"
				note = "some note"
				url_parts = {
					scheme = "http"
					host = "my-totally-random-custom-host.com"
					port = 12345
					base_path = "some-dir"
				}
			}`,
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionNoop),
		},
		resource.TestStep{
			Config: `resource "cdn77_origin_url" "url" {
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
				checkUrl(client, &originId, func(o *cdn77.UrlOriginDetail) error {
					return errors.Join(
						acctest.EqualField("type", o.Type, origin.TypeUrl),
						acctest.EqualField("label", o.Label, "another label"),
						acctest.NullField("note", o.Note),
						acctest.EqualField("scheme", o.Scheme, "http"),
						acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
						acctest.NullField("port", o.Port),
						acctest.NullField("base_dir", o.BaseDir),
					)
				}),
			),
		},
	)
}

func TestAccOrigin_UrlResource_Import(t *testing.T) {
	client := acctest.GetClient(t)
	rsc := "cdn77_origin_url.url"
	var originId string

	acctest.Run(t, acctest.CheckOriginDestroyed(client, origin.TypeUrl),
		resource.TestStep{
			Config: `resource "cdn77_origin_url" "url" {
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
				return originId, nil
			},
			ImportStateVerify: true,
		},
	)
}

func TestAccOrigin_UrlDataSource_OnlyRequiredFields(t *testing.T) {
	const nonExistingOriginId = "bcd7b5bb-a044-4611-82e4-3f3b2a3cda13"
	const rsc = "data.cdn77_origin_url.url"
	const label = "random origin"
	const originUrl = "http://my-totally-random-custom-host.com"
	const scheme = "http"
	const host = "my-totally-random-custom-host.com"
	client := acctest.GetClient(t)
	request := cdn77.OriginCreateUrlJSONRequestBody{
		Label:  label,
		Scheme: scheme,
		Host:   host,
	}

	response, err := client.OriginCreateUrlWithResponse(t.Context(), request)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response, err)

	originId := response.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, origin.TypeUrl, originId)
	})

	acctest.Run(t, nil,
		resource.TestStep{
			Config:      acctest.Config(urlDataSourceConfig, "id", nonExistingOriginId),
			ExpectError: regexp.MustCompile(fmt.Sprintf(`.*?"%s".*?not found.*?`, nonExistingOriginId)),
		},
		resource.TestStep{
			Config: acctest.Config(urlDataSourceConfig, "id", originId),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(rsc, "id", originId),
				resource.TestCheckResourceAttr(rsc, "label", label),
				resource.TestCheckResourceAttr(rsc, "url", originUrl),
				resource.TestCheckResourceAttr(rsc, "url_parts.scheme", scheme),
				resource.TestCheckResourceAttr(rsc, "url_parts.host", host),
				resource.TestCheckNoResourceAttr(rsc, "note"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.port"),
				resource.TestCheckNoResourceAttr(rsc, "url_parts.base_path"),
			),
		},
	)
}

func TestAccOrigin_UrlDataSource_AllFields(t *testing.T) {
	const rsc = "data.cdn77_origin_url.url"
	const label = "random origin"
	const note = "some note"
	const originUrl = "https://my-totally-random-custom-host.com:12345/some-dir"
	const scheme = "https"
	const host = "my-totally-random-custom-host.com"
	const port = 12345
	const basePath = "some-dir"
	client := acctest.GetClient(t)
	request := cdn77.OriginCreateUrlJSONRequestBody{
		Label:   label,
		Note:    nullable.NewNullableWithValue(note),
		Scheme:  scheme,
		Host:    host,
		Port:    nullable.NewNullableWithValue(port),
		BaseDir: nullable.NewNullableWithValue(basePath),
	}

	response, err := client.OriginCreateUrlWithResponse(t.Context(), request)
	acctest.AssertResponseOk(t, "Failed to create Origin: %s", response, err)

	originId := response.JSON201.Id

	t.Cleanup(func() {
		acctest.MustDeleteOrigin(t, client, origin.TypeUrl, originId)
	})

	acctest.Run(t, nil, resource.TestStep{
		Config: acctest.Config(urlDataSourceConfig, "id", originId),
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(rsc, "id", originId),
			resource.TestCheckResourceAttr(rsc, "label", label),
			resource.TestCheckResourceAttr(rsc, "note", note),
			resource.TestCheckResourceAttr(rsc, "url", originUrl),
			resource.TestCheckResourceAttr(rsc, "url_parts.scheme", scheme),
			resource.TestCheckResourceAttr(rsc, "url_parts.host", host),
			resource.TestCheckResourceAttr(rsc, "url_parts.port", strconv.Itoa(port)),
			resource.TestCheckResourceAttr(rsc, "url_parts.base_path", basePath),
		),
	})
}

func checkUrl(
	client cdn77.ClientWithResponsesInterface,
	originId *string,
	fn func(o *cdn77.UrlOriginDetail) error,
) func(*terraform.State) error {
	return func(*terraform.State) error {
		response, err := client.OriginDetailUrlWithResponse(context.Background(), *originId)
		message := fmt.Sprintf("failed to get Origin[id=%s]: %%s", *originId)

		if err = acctest.CheckResponse(message, response, err); err != nil {
			return err
		}

		return fn(response.JSON200)
	}
}

const urlDataSourceConfig = `
data "cdn77_origin_url" "url" {
  id = "{id}"
}
`
