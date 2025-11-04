package ssl

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type noLeadingTrailingWhitespaceValidator struct{}

func (noLeadingTrailingWhitespaceValidator) Description(_ context.Context) string {
	return "value must not have leading or trailing whitespace"
}

func (noLeadingTrailingWhitespaceValidator) MarkdownDescription(_ context.Context) string {
	return "value must not have leading or trailing whitespace. Use `trimspace()` function to remove whitespace."
}

func (noLeadingTrailingWhitespaceValidator) ValidateString(
	_ context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Empty Value",
			"Certificate/private key must not be empty. If you load from a file, ensure it contains a valid PEM block.",
		)

		return
	}

	if value != trimmed {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Value",
			"Certificate and private key values must not have leading or trailing whitespace. "+
				"Use trimspace(file(\"${path.module}/cert.pem\")) or chomp(file(\"${path.module}/cert.pem\"))",
		)
	}
}

func NoLeadingTrailingWhitespace() validator.String {
	return noLeadingTrailingWhitespaceValidator{}
}
