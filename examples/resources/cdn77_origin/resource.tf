resource "cdn77_origin" "example" {
  type   = "url"
  label  = "Static content for example.com"
  scheme = "http"
  host   = "static.example.com"
}
