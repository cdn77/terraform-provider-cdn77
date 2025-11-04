package ssl_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest/testdata"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSslResource(t *testing.T) {
	const rsc = "cdn77_ssl.crt"
	client := acctest.GetClient(t)
	var sslId string

	acctest.Run(t, checkSslsDestroyed(client),
		resource.TestStep{
			Config:           acctest.Config(resourceConfig, "cert", testdata.SslCert1, "key", testdata.SslKey),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionCreate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAndAssignAttr(rsc, "id", &sslId),
				resource.TestCheckResourceAttr(rsc, "certificate", testdata.SslCert1),
				resource.TestCheckResourceAttr(rsc, "private_key", testdata.SslKey),
				resource.TestCheckResourceAttr(rsc, "subjects.#", "2"),
				resource.TestCheckTypeSetElemAttr(rsc, "subjects.*", "cdn.example.com"),
				resource.TestCheckTypeSetElemAttr(rsc, "subjects.*", "other.mycdn.cz"),
				resource.TestCheckResourceAttr(rsc, "expires_at", "2051-09-02 12:19:20"),

				checkSsl(client, &sslId, func(o *cdn77.Ssl) error {
					sort.Stable(sort.StringSlice(o.Cnames))

					return errors.Join(
						acctest.EqualField("certificate", o.Certificate, testdata.SslCert1),
						acctest.EqualField("expires_at", o.ExpiresAt.Format(time.DateTime), "2051-09-02 12:19:20"),
						acctest.EqualField("len(cnames)", len(o.Cnames), 2),
						acctest.EqualField("cnames.0", o.Cnames[0], "cdn.example.com"),
						acctest.EqualField("cnames.1", o.Cnames[1], "other.mycdn.cz"),
					)
				}),
			),
		},
		resource.TestStep{
			Config:           acctest.Config(resourceConfig, "cert", testdata.SslCert2, "key", testdata.SslKey),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionUpdate),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAttr(rsc, "id", &sslId),
				resource.TestCheckResourceAttr(rsc, "certificate", testdata.SslCert2),
				resource.TestCheckResourceAttr(rsc, "private_key", testdata.SslKey),
				resource.TestCheckResourceAttr(rsc, "subjects.#", "1"),
				resource.TestCheckTypeSetElemAttr(rsc, "subjects.*", "mycdn.cz"),
				resource.TestCheckResourceAttr(rsc, "expires_at", "2051-09-03 07:51:29"),

				checkSsl(client, &sslId, func(o *cdn77.Ssl) error {
					sort.Stable(sort.StringSlice(o.Cnames))

					return errors.Join(
						acctest.EqualField("certificate", o.Certificate, testdata.SslCert2),
						acctest.EqualField("expires_at", o.ExpiresAt.Format(time.DateTime), "2051-09-03 07:51:29"),
						acctest.EqualField("len(cnames)", len(o.Cnames), 1),
						acctest.EqualField("cnames.0", o.Cnames[0], "mycdn.cz"),
					)
				}),
			),
		},
	)
}

func TestAccSslResource_Import(t *testing.T) {
	client := acctest.GetClient(t)
	rsc := "cdn77_ssl.crt"
	var sslId string

	acctest.Run(t, checkSslsDestroyed(client),
		resource.TestStep{
			Config: acctest.Config(resourceConfig, "cert", testdata.SslCert1, "key", testdata.SslKey),
			Check:  acctest.CheckAndAssignAttr(rsc, "id", &sslId),
		},
		resource.TestStep{
			ResourceName:      rsc,
			ImportState:       true,
			ImportStateVerify: true,
			ImportStateIdFunc: func(*terraform.State) (string, error) {
				privateKeyBase64 := base64.StdEncoding.EncodeToString([]byte(testdata.SslKey))

				return fmt.Sprintf("%s,%s", sslId, privateKeyBase64), nil
			},
		},
	)
}

func TestAccSslResource_WhitespaceHandling(t *testing.T) {
	const rsc = "cdn77_ssl.crt"
	client := acctest.GetClient(t)
	var sslId string

	certWithWhitespace := "\n  " + testdata.SslCert1 + "  \n"
	keyWithWhitespace := "\n  " + testdata.SslKey + "  \n"

	acctest.Run(t, checkSslsDestroyed(client),
		resource.TestStep{
			Config: acctest.Config(resourceConfig, "cert", certWithWhitespace, "key", keyWithWhitespace),
			Check: resource.ComposeAggregateTestCheckFunc(
				acctest.CheckAndAssignAttr(rsc, "id", &sslId),
				resource.TestCheckResourceAttr(rsc, "certificate", testdata.SslCert1),
				checkSsl(client, &sslId, func(o *cdn77.Ssl) error {
					return acctest.EqualField("certificate", o.Certificate, testdata.SslCert1)
				}),
			),
		},
		resource.TestStep{
			Config:           acctest.Config(resourceConfig, "cert", certWithWhitespace, "key", keyWithWhitespace),
			ConfigPlanChecks: acctest.ConfigPlanChecks(rsc, plancheck.ResourceActionNoop),
		},
	)
}

func TestAccSslResource_FileWithoutChomp(t *testing.T) {
	certFile := createTempFile(t, "\n  "+testdata.SslCert1+"  \n")
	keyFile := createTempFile(t, "\n  "+testdata.SslKey+"  \n")

	configWithFile := fmt.Sprintf(`
resource "cdn77_ssl" "crt" {
 certificate = file("%s")
 private_key = file("%s")
}`, certFile, keyFile)

	expectedError := `(?s)Certificate and private key values must not have leading or trailing\s+whitespace`
	acctest.Run(t, nil,
		resource.TestStep{
			Config:      configWithFile,
			ExpectError: regexp.MustCompile(expectedError),
		},
	)
}

func TestAccSslDataSource(t *testing.T) {
	const nonExistingSslId = "ae4f471f-029a-4a5c-b9bd-27ea28815de0"
	client := acctest.GetClient(t)
	sslId := acctest.MustAddSslWithCleanup(t, client, testdata.SslCert1, testdata.SslKey)

	acctest.Run(t, nil,
		resource.TestStep{
			Config: acctest.Config(dataSourceConfig, "id", nonExistingSslId),
			ExpectError: regexp.MustCompile(
				fmt.Sprintf(`SNI certificate with id "%s" was not\s+found.`, nonExistingSslId),
			),
		},
		resource.TestStep{
			Config: acctest.Config(dataSourceConfig, "id", sslId),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.cdn77_ssl.ipsum", "id", sslId),
				resource.TestCheckResourceAttr("data.cdn77_ssl.ipsum", "certificate", testdata.SslCert1),
				resource.TestCheckResourceAttr("data.cdn77_ssl.ipsum", "subjects.#", "2"),
				resource.TestCheckTypeSetElemAttr("data.cdn77_ssl.ipsum", "subjects.*", "cdn.example.com"),
				resource.TestCheckTypeSetElemAttr("data.cdn77_ssl.ipsum", "subjects.*", "other.mycdn.cz"),
				resource.TestCheckResourceAttr("data.cdn77_ssl.ipsum", "expires_at", "2051-09-02 12:19:20"),
				resource.TestCheckNoResourceAttr("data.cdn77_ssl.ipsum", "private_key"),
			),
		},
	)
}

func checkSsl(
	client cdn77.ClientWithResponsesInterface,
	sslId *string,
	fn func(o *cdn77.Ssl) error,
) func(*terraform.State) error {
	return func(*terraform.State) error {
		response, err := client.SslSniDetailWithResponse(context.Background(), *sslId)
		message := fmt.Sprintf("failed to get SSL[id=%s]: %%s", *sslId)

		if err = acctest.CheckResponse(message, response, err); err != nil {
			return err
		}

		return fn(response.JSON200)
	}
}

func checkSslsDestroyed(client cdn77.ClientWithResponsesInterface) resource.TestCheckFunc {
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

func createTempFile(t *testing.T, content string) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "test-*.pem")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.WriteString(content); err != nil {
		t.Fatal(err)
	}

	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	return file.Name()
}

const resourceConfig = `
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

const dataSourceConfig = `
data "cdn77_ssl" "ipsum" {
  id = "{id}"
}
`
