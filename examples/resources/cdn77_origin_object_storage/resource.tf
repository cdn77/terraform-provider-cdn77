data "cdn77_object_storages" "all" {
}

locals {
  eu_cluster_id = one([for os in data.cdn77_object_storages.all.clusters : os.id if os.label == "EU"])
}

resource "cdn77_origin_object_storage" "example" {
  label             = "Assets bucket for example.com"
  bucket_name       = "examplecom-static-assets"
  acl               = "private"
  cluster_id        = local.eu_cluster_id
  access_key_id     = "I17DXFE00GNJZVQUTQPW"
  access_key_secret = "7UG7WbcIz4VhZnVxV4XQcDR2X0APApuvthyATf2v"
}
