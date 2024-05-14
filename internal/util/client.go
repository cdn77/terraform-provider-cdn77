package util

import (
	"fmt"

	"github.com/cdn77/cdn77-client-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func MaybeSetClient(providerData any, target *cdn77.ClientWithResponsesInterface) diag.Diagnostic {
	if providerData == nil {
		return nil
	}

	client, ok := providerData.(cdn77.ClientWithResponsesInterface)
	if !ok {
		const detail = "Expected cdn77.ClientWithResponsesInterface, got: %T. " +
			"Please report this issue to the provider developers."

		return diag.NewErrorDiagnostic("Unexpected provider data type", fmt.Sprintf(detail, providerData))
	}

	*target = client

	return nil
}
