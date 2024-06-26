---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cdn77 Provider"
subcategory: ""
description: |-
  The CDN77 provider is used to interact with CDN77 resources.
---

# cdn77 Provider

The CDN77 provider is used to interact with CDN77 resources.

## Example Usage

```terraform
provider "cdn77" {
  # Required only if the CDN77_TOKEN env variable isn't set
  token = "--- your secret token here ---"

  # Optional fields with their default values
  endpoint = "https://api.cdn77.com"
  timeout  = 10 # in seconds
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `endpoint` (String) API endpoint; defaults to https://api.cdn77.com
- `timeout` (Number) Timeout for all API calls (in seconds). Negative values disable the timeout. Default is 30 seconds.
- `token` (String, Sensitive) Authentication token from https://client.cdn77.com/account/api
