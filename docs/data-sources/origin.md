---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cdn77_origin Data Source - terraform-provider-cdn77"
subcategory: ""
description: |-
  Origin resource allows you to manage your Origins
---

# cdn77_origin (Data Source)

Origin resource allows you to manage your Origins

## Example Usage

```terraform
data "cdn77_origin" "example" {
  id   = "2a81317a-53f4-4c6b-b3ca-9c1f9bc0ac41"
  type = "aws"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) Origin ID (UUID)
- `type` (String) Type of the origin; one of [aws object-storage url]

### Read-Only

- `access_key_id` (String) Access key to your Object Storage bucket
- `access_key_secret` (String, Sensitive) Access secret to your Object Storage bucket
- `acl` (String) Object Storage access key ACL
- `aws_access_key_id` (String) AWS access key ID
- `aws_access_key_secret` (String, Sensitive) AWS access key secret
- `aws_region` (String) AWS region
- `base_dir` (String) Directory where the content is stored on the URL Origin
- `bucket_name` (String) Name of your Object Storage bucket
- `cluster_id` (String) ID of the Object Storage storage cluster
- `host` (String) Origin host without scheme and port. Can be domain name or IP address
- `label` (String) The label helps you to identify your Origin
- `note` (String) Optional note for the Origin
- `port` (Number) Origin port number. If not specified, default scheme port is used
- `scheme` (String) Scheme of the Origin
