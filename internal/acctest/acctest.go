package acctest

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func GetProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"cdn77": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func GetClient(t *testing.T) cdn77.ClientWithResponsesInterface {
	t.Helper()

	client, err := GetClientErr()
	if err != nil {
		t.Fatal(err.Error())
	}

	return client
}

func GetClientErr() (cdn77.ClientWithResponsesInterface, error) {
	endpoint := os.Getenv("CDN77_ENDPOINT")
	token := os.Getenv("CDN77_TOKEN")
	var timeout time.Duration

	if timeoutString := os.Getenv("CDN77_TIMEOUT"); timeoutString != "" {
		timeoutSeconds, err := strconv.Atoi(timeoutString)
		if err != nil {
			const message = `CDN77_TIMEOUT contains invalid timeout "%s"; expected integer (number of seconds)`

			return nil, fmt.Errorf(message, timeoutString)
		}

		timeout = time.Second * time.Duration(timeoutSeconds)
	}

	if endpoint == "" {
		endpoint = "https://api.cdn77.com"
	}

	if token == "" {
		return nil, errors.New("CDN77_TOKEN must be set for acceptance tests")
	}

	if timeout == 0 {
		timeout = time.Second * 10
	}

	client, err := provider.NewClient(endpoint, token, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize API client: %w", err)
	}

	return client, nil
}

func AssertResponseOk(t *testing.T, message string, response any, err error) {
	t.Helper()

	if err := CheckResponse(message, response, err); err != nil {
		t.Fatal(err.Error())
	}
}

func CheckResponse(message string, response any, err error) error {
	vResponse := reflect.Indirect(reflect.ValueOf(response))

	if err != nil {
		return fmt.Errorf(message, err)
	}

	statusCodeMethod := vResponse.MethodByName("StatusCode")
	if !statusCodeMethod.IsValid() {
		return fmt.Errorf(message, "missing StatusCode method on the response object")
	}

	values := statusCodeMethod.Call(nil)
	if len(values) != 1 || !values[0].CanInt() {
		return fmt.Errorf(message, "unexpected StatusCode method signature")
	}

	statusCode := int(values[0].Int())
	if statusCode >= 200 && statusCode <= 204 {
		return nil
	}

	var body string
	if vBody := vResponse.FieldByName("Body"); vBody.IsValid() {
		body = string(vBody.Bytes())
	}

	return fmt.Errorf(message, fmt.Sprintf("unexpected HTTP status code: %d; response: %s", statusCode, body))
}

func Config(config string, keyAndValues ...any) string {
	if len(keyAndValues)%2 != 0 {
		panic("keyAndValues must be pairs")
	}

	for i := 0; i < len(keyAndValues); i += 2 {
		key := keyAndValues[i].(string)
		var valueString string

		switch value := keyAndValues[i+1].(type) {
		case int:
			valueString = strconv.Itoa(value)
		case string:
			valueString = value
		default:
			panic(fmt.Sprintf("unknown value type: %T", value))
		}

		config = strings.ReplaceAll(config, fmt.Sprintf("{%s}", key), valueString)
	}

	return config
}
