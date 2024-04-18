package util

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/cdn77/cdn77-client-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oapi-codegen/nullable"
)

type StatusCodeProvider interface {
	StatusCode() int
}

func IntPointerToInt64Value[T ~int](v *T) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}

	return types.Int64Value(int64(*v))
}

func Int64ValueToNullable[T ~int](v types.Int64) nullable.Nullable[T] {
	if v.IsNull() {
		return nullable.NewNullNullable[T]()
	}

	return nullable.NewNullableWithValue(T(v.ValueInt64()))
}

func StringValueToNullable(v types.String) nullable.Nullable[string] {
	if v.IsNull() {
		return nullable.NewNullNullable[string]()
	}

	return nullable.NewNullableWithValue(v.ValueString())
}

func NullableIntToInt64Value[T ~int](v nullable.Nullable[T]) types.Int64 {
	if v.IsNull() || !v.IsSpecified() {
		return types.Int64Null()
	}

	return types.Int64Value(int64(v.MustGet()))
}

func NullableToStringValue(v nullable.Nullable[string]) types.String {
	if v.IsNull() || !v.IsSpecified() {
		return types.StringNull()
	}

	return types.StringValue(v.MustGet())
}

func Pointer[T any](v T) *T {
	return &v
}

func CheckResponse(diags *diag.Diagnostics, message string, response StatusCodeProvider, errMessages ...any) bool {
	for _, errMessage := range errMessages {
		if reflect.ValueOf(errMessage).IsNil() {
			continue
		}

		var detail string
		switch m := errMessage.(type) {
		case *cdn77.Errors:
			detail = buildResponseErrMessage(response.StatusCode(), m.Errors, nil)
		case *cdn77.FieldErrors:
			detail = buildResponseErrMessage(response.StatusCode(), m.Errors, m.Fields)
		default:
			panic(fmt.Sprintf(`unexpected error response type "%T" (HTTP %d)`, errMessage, response.StatusCode()))
		}

		diags.AddError(message, detail)

		return false
	}

	return true
}

func buildResponseErrMessage(statusCode int, errs []string, fields map[string][]string) string {
	var fieldsMessage string
	if fieldsCount := len(fields); fieldsCount != 0 {
		fieldsMessage = "\n\nFollowing fields have some errors:\n"

		for field, errs := range fields {
			switch len(errs) {
			case 0:
				fieldsMessage += fmt.Sprintf("\t%s: <unknown>\n", field)
			case 1:
				fieldsMessage += fmt.Sprintf("\t%s: %s\n", field, errs[0])
			default:
				fieldsMessage += fmt.Sprintf("\t%s:\n\t\t%s\n", field, strings.Join(errs, "\n\t\t"))
			}
		}
	}

	httpErr := fmt.Sprintf("HTTP %d %s", statusCode, http.StatusText(statusCode))

	switch len(errs) {
	case 0:
		return fmt.Sprintf("Received unexpected API response:\n\t%s%s", httpErr, fieldsMessage)
	case 1:
		return fmt.Sprintf("Received unexpected API error:\n\t%s%s\n\n%s", errs[0], fieldsMessage, httpErr)
	default:
		return fmt.Sprintf(
			"Received unexpected API errors:\n\t%s%s\n\n%s",
			strings.Join(errs, "\n\t"),
			fieldsMessage,
			httpErr,
		)
	}
}
