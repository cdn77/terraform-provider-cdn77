package acctest

import (
	"context"
	"fmt"
	"testing"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider"
)

func DeleteOrigin(client cdn77.ClientWithResponsesInterface, originType string, id string) (err error) {
	var response any

	switch originType {
	case provider.OriginTypeAws:
		response, err = client.OriginDeleteAwsWithResponse(context.Background(), id)
	case provider.OriginTypeObjectStorage:
		response, err = client.OriginDeleteObjectStorageWithResponse(context.Background(), id)
	case provider.OriginTypeUrl:
		response, err = client.OriginDeleteUrlWithResponse(context.Background(), id)
	default:
		panic(fmt.Sprintf("unknown Origin type: %s", originType))
	}

	return CheckResponse(fmt.Sprintf("failed to delete Origin[id=%s]: %%s", id), response, err)
}

func MustDeleteOrigin(t *testing.T, client cdn77.ClientWithResponsesInterface, originType string, id string) {
	t.Helper()

	if err := DeleteOrigin(client, originType, id); err != nil {
		t.Fatal(err.Error())
	}
}

func DeleteCdn(client cdn77.ClientWithResponsesInterface, id int) error {
	response, err := client.CdnDeleteWithResponse(context.Background(), id)

	return CheckResponse(fmt.Sprintf("failed to delete CDN[id=%d]: %%s", id), response, err)
}

func MustDeleteCdn(t *testing.T, client cdn77.ClientWithResponsesInterface, id int) {
	t.Helper()

	if err := DeleteCdn(client, id); err != nil {
		t.Fatal(err.Error())
	}
}

func DeleteSsl(client cdn77.ClientWithResponsesInterface, id string) error {
	response, err := client.SslSniDeleteWithResponse(context.Background(), id)

	return CheckResponse(fmt.Sprintf("failed to delete SSL[id=%s]: %%s", id), response, err)
}

func MustDeleteSsl(t *testing.T, client cdn77.ClientWithResponsesInterface, id string) {
	t.Helper()

	if err := DeleteSsl(client, id); err != nil {
		t.Fatal(err.Error())
	}
}

func MustAddSslWithCleanup(t *testing.T, client cdn77.ClientWithResponsesInterface, certAndKey ...string) string {
	t.Helper()

	if len(certAndKey)%2 != 0 {
		t.Fatal("certAndKey must be a pair")
	}

	sslRequest := cdn77.SslSniAddJSONRequestBody{Certificate: certAndKey[0], PrivateKey: certAndKey[1]}
	sslResponse, err := client.SslSniAddWithResponse(context.Background(), sslRequest)
	AssertResponseOk(t, "Failed to add SSL: %s", sslResponse, err)

	sslId := sslResponse.JSON201.Id

	t.Cleanup(func() {
		MustDeleteSsl(t, client, sslId)
	})

	return sslId
}
