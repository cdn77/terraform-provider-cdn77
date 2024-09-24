package util

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oapi-codegen/nullable"
)

type Response interface {
	StatusCode() int
	Bytes() []byte
}

func IntPointerToInt64Value[T ~int](v *T) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}

	return types.Int64Value(int64(*v))
}

func Int64ValueToNullable[T ~int](v types.Int64) nullable.Nullable[T] {
	if v.IsNull() || v.IsUnknown() {
		return nullable.NewNullNullable[T]()
	}

	return nullable.NewNullableWithValue(T(v.ValueInt64()))
}

func Int32ValueToNullable[T ~int](v types.Int32) nullable.Nullable[T] {
	if v.IsNull() || v.IsUnknown() {
		return nullable.NewNullNullable[T]()
	}

	return nullable.NewNullableWithValue(T(v.ValueInt32()))
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

func NullableIntToInt32Value[T ~int](v nullable.Nullable[T]) types.Int32 {
	if v.IsNull() || !v.IsSpecified() {
		return types.Int32Null()
	}

	return types.Int32Value(int32(v.MustGet()))
}

func NullableToStringValue(v nullable.Nullable[string]) types.String {
	if v.IsNull() || !v.IsSpecified() {
		return types.StringNull()
	}

	return types.StringValue(v.MustGet())
}

func ProcessResponse[T any](
	diags *diag.Diagnostics,
	response Response,
	errMessage string,
	okResponse *T,
	fn func(*T),
) {
	if !processResponseErrors(diags, response, errMessage) {
		return
	}

	if okResponse == nil {
		unexpectedApiError(diags, response, errMessage)

		return
	}

	fn(okResponse)
}

func ProcessEmptyResponse(diags *diag.Diagnostics, response Response, errMessage string, fn func()) {
	if !processResponseErrors(diags, response, errMessage) {
		return
	}

	if response.StatusCode() != http.StatusNoContent {
		unexpectedApiError(diags, response, errMessage)

		return
	}

	fn()
}

func ValidateDeletionResponse(diags *diag.Diagnostics, response Response, errMessage string) {
	if response.StatusCode() == http.StatusNotFound {
		return
	}

	ProcessEmptyResponse(diags, response, errMessage, func() {})
}

func processResponseErrors(diags *diag.Diagnostics, response Response, errMessage string) bool {
	vResponse := reflect.Indirect(reflect.ValueOf(response))
	tResponse := vResponse.Type()

	for i := tResponse.NumField() - 1; i >= 0; i-- {
		tField := tResponse.Field(i)

		name, ok := strings.CutPrefix(tField.Name, "JSON")
		if !ok {
			continue
		}

		if status, err := strconv.Atoi(name); err != nil && name != "Default" || status < 300 {
			continue
		}

		vField := vResponse.Field(i)
		if vField.IsNil() {
			continue
		}

		var detail string
		switch m := vField.Interface().(type) {
		case *cdn77.Errors:
			detail = buildResponseErrMessage(response, m.Errors, nil)
		case *cdn77.FieldErrors:
			detail = buildResponseErrMessage(response, m.Errors, m.Fields)
		default:
			detail = fmt.Sprintf(
				"Unexpected error response type \"%T\"\nHTTP %d %s\n\n%s\n",
				vField.Interface(),
				response.StatusCode(),
				http.StatusText(response.StatusCode()),
				response.Bytes(),
			)
		}

		diags.AddError(errMessage, detail)

		return false
	}

	return true
}

func unexpectedApiError(diags *diag.Diagnostics, response Response, errMessage string) {
	code := response.StatusCode()
	detail := fmt.Sprintf("Unexpected API response\nHTTP %d %s\n\n%s\n", code, http.StatusText(code), response.Bytes())

	diags.AddError(errMessage, detail)
}

func buildResponseErrMessage(response Response, errs []string, fields map[string][]string) string {
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

	httpErr := fmt.Sprintf("HTTP %d %s", response.StatusCode(), http.StatusText(response.StatusCode()))

	switch len(errs) {
	case 0:
		if fieldsMessage == "" {
			return fmt.Sprintf("Received unexpected API response:\n\t%s\n%s", httpErr, response.Bytes())
		}

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
