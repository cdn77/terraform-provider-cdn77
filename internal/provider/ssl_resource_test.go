package provider_test

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSslResource(t *testing.T) {
	client := acctest.GetClient(t)
	var sslId string

	rsc := "cdn77_ssl.crt"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		CheckDestroy:             checkSslsDestroyed(client),
		Steps: []resource.TestStep{
			{
				Config: acctest.Config(SslResourceConfig, "cert", sslTestCert1, "key", sslTestKey),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(rsc, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(rsc, "id", func(value string) error {
						sslId = value

						return acctest.NotEqual(value, "")
					}),
					resource.TestCheckResourceAttr(rsc, "certificate", sslTestCert1),
					resource.TestCheckResourceAttr(rsc, "private_key", sslTestKey),
					resource.TestCheckResourceAttr(rsc, "subjects.#", "2"),
					resource.TestCheckTypeSetElemAttr(rsc, "subjects.*", "cdn.example.com"),
					resource.TestCheckTypeSetElemAttr(rsc, "subjects.*", "other.mycdn.cz"),
					resource.TestCheckResourceAttr(rsc, "expires_at", "2051-09-02 12:19:20"),

					checkSsl(client, &sslId, func(o *cdn77.Ssl) error {
						sort.Stable(sort.StringSlice(o.Cnames))

						return errors.Join(
							acctest.EqualField("certificate", o.Certificate, sslTestCert1),
							acctest.EqualField("expires_at", o.ExpiresAt.Format(time.DateTime), "2051-09-02 12:19:20"),
							acctest.EqualField("len(cnames)", len(o.Cnames), 2),
							acctest.EqualField("cnames.0", o.Cnames[0], "cdn.example.com"),
							acctest.EqualField("cnames.1", o.Cnames[1], "other.mycdn.cz"),
						)
					}),
				),
			},
			{
				Config: acctest.Config(SslResourceConfig, "cert", sslTestCert2, "key", sslTestKey),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(rsc, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(rsc, "id", func(value string) error {
						return acctest.Equal(value, sslId)
					}),
					resource.TestCheckResourceAttr(rsc, "certificate", sslTestCert2),
					resource.TestCheckResourceAttr(rsc, "private_key", sslTestKey),
					resource.TestCheckResourceAttr(rsc, "subjects.#", "1"),
					resource.TestCheckTypeSetElemAttr(rsc, "subjects.*", "mycdn.cz"),
					resource.TestCheckResourceAttr(rsc, "expires_at", "2051-09-03 07:51:29"),

					checkSsl(client, &sslId, func(o *cdn77.Ssl) error {
						sort.Stable(sort.StringSlice(o.Cnames))

						return errors.Join(
							acctest.EqualField("certificate", o.Certificate, sslTestCert2),
							acctest.EqualField("expires_at", o.ExpiresAt.Format(time.DateTime), "2051-09-03 07:51:29"),
							acctest.EqualField("len(cnames)", len(o.Cnames), 1),
							acctest.EqualField("cnames.0", o.Cnames[0], "mycdn.cz"),
						)
					}),
				),
			},
		},
	})
}

func checkSsl(
	client cdn77.ClientWithResponsesInterface,
	sslId *string,
	fn func(o *cdn77.Ssl) error,
) func(*terraform.State) error {
	return func(_ *terraform.State) error {
		response, err := client.SslSniDetailWithResponse(context.Background(), *sslId)
		message := fmt.Sprintf("failed to get SSL[id=%s]: %%s", *sslId)

		if err = acctest.CheckResponse(message, response, err); err != nil {
			return err
		}

		return fn(response.JSON200)
	}
}

func checkSslsDestroyed(client cdn77.ClientWithResponsesInterface) func(*terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "cdn77_ssl" {
				continue
			}

			response, err := client.SslSniDetailWithResponse(context.Background(), rs.Primary.Attributes["id"])
			if err != nil {
				return fmt.Errorf("failed to fetch SSL: %w", err)
			}

			if response.JSON404 == nil {
				return errors.New("expected SSL to be deleted")
			}
		}

		return nil
	}
}

const SslResourceConfig = `
resource "cdn77_ssl" "crt" {
	certificate = trimspace(
	<<EOT
		{cert}
	EOT
	)
	private_key = trimspace(
	<<EOT
		{key}
	EOT
	)
}
`
