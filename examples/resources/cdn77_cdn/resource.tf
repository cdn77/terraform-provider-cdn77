resource "cdn77_cdn" "example" {
  label     = "Static content for example.com"
  origin_id = cdn77_origin_url.example.id
  cnames    = ["cdn.example.com"]
}
