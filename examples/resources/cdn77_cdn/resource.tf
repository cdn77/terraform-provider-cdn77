resource "cdn77_cdn" "example" {
  origin_id = cdn77_origin.example.id
  label     = "Static content for example.com"
  cnames    = ["cdn.example.com"]
}
