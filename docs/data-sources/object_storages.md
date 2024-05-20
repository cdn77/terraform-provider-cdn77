---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cdn77_object_storages Data Source - terraform-provider-cdn77"
subcategory: ""
description: |-
  Object Storages data source allows you to read all available Object Storage clusters
---

# cdn77_object_storages (Data Source)

Object Storages data source allows you to read all available Object Storage clusters



<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `clusters` (Attributes List) List of all Object Storage clusters (see [below for nested schema](#nestedatt--clusters))

<a id="nestedatt--clusters"></a>
### Nested Schema for `clusters`

Read-Only:

- `host` (String)
- `id` (String) ID (UUID) of the Object Storage cluster
- `label` (String)
- `port` (Number)
- `scheme` (String)