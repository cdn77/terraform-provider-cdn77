package acctest

import (
	"context"
	"errors"
	"fmt"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/origin"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func CheckOriginDestroyed(client cdn77.ClientWithResponsesInterface, originType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			var resourceTypeName string

			switch originType {
			case origin.TypeAws:
				resourceTypeName = "cdn77_origin_aws"
			case origin.TypeObjectStorage:
				resourceTypeName = "cdn77_origin_object_storage"
			case origin.TypeUrl:
				resourceTypeName = "cdn77_origin_url"
			default:
				return fmt.Errorf("unknown Origin type: %s", originType)
			}

			if rs.Type != resourceTypeName {
				continue
			}

			response404, err := getOriginDetail(client, originType, rs.Primary.Attributes["id"])
			if err != nil {
				return err
			}

			if response404 == nil {
				return errors.New("expected origin to be deleted")
			}
		}

		return nil
	}
}

func getOriginDetail(client cdn77.ClientWithResponsesInterface, originType string, id string) (*cdn77.Errors, error) {
	switch originType {
	case origin.TypeAws:
		response, err := client.OriginDetailAwsWithResponse(context.Background(), id)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch AWS Origin: %w", err)
		}

		return response.JSON404, nil
	case origin.TypeObjectStorage:
		response, err := client.OriginDetailObjectStorageWithResponse(context.Background(), id)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch Object Storage Origin: %w", err)
		}

		return response.JSON404, nil
	case origin.TypeUrl:
		response, err := client.OriginDetailUrlWithResponse(context.Background(), id)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch URL Origin: %w", err)
		}

		return response.JSON404, nil
	default:
		return nil, fmt.Errorf("unknown Origin type: %s", originType)
	}
}
