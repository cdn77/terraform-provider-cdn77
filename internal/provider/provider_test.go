package provider_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMain(m *testing.M) {
	resource.AddTestSweepers("cdn77_cdn", &resource.Sweeper{
		Name: "cdn77_cdn",
		F: func(_ string) error {
			client, err := acctest.GetClientErr()
			if err != nil {
				return err
			}

			response, err := client.CdnListWithResponse(context.Background())
			if err = acctest.CheckResponse("failed to list CDNs: %s", response, err); err != nil {
				return err
			}

			for _, cdn := range *response.JSON200 {
				response, err := client.CdnDeleteWithResponse(context.Background(), cdn.Id)
				message := fmt.Sprintf("failed to delete CDN[id=%d]: %%s", cdn.Id)

				if err = acctest.CheckResponse(message, response, err); err != nil {
					return err
				}
			}

			return nil
		},
	})
	resource.AddTestSweepers("cdn77_origin", &resource.Sweeper{
		Name:         "cdn77_origin",
		Dependencies: []string{"cdn77_cdn"},
		F: func(_ string) error {
			client, err := acctest.GetClientErr()
			if err != nil {
				return err
			}

			response, err := client.OriginListWithResponse(context.Background())
			if err = acctest.CheckResponse("failed to list Origins: %s", response, err); err != nil {
				return err
			}

			for _, originListItem := range *response.JSON200 {
				origin, err := originListItem.AsUrlOriginDetail()
				if err != nil {
					panic(err)
				}

				if err = acctest.DeleteOrigin(client, string(origin.Type), origin.Id); err != nil {
					return err
				}
			}

			return nil
		},
	})
	resource.AddTestSweepers("cdn77_ssl", &resource.Sweeper{
		Name:         "cdn77_ssl",
		Dependencies: []string{"cdn77_cdn"},
		F: func(_ string) error {
			client, err := acctest.GetClientErr()
			if err != nil {
				return err
			}

			response, err := client.SslSniListWithResponse(context.Background())
			if err = acctest.CheckResponse("failed to list SSL SNIs: %s", response, err); err != nil {
				return err
			}

			for _, sslSni := range *response.JSON200 {
				if err = acctest.DeleteSsl(client, sslSni.Id); err != nil {
					return err
				}
			}

			return nil
		},
	})

	resource.TestMain(m)
}
