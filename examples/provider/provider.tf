provider "cdn77" {
  # Required only if the CDN77_TOKEN env variable isn't set
  token = "--- your secret token here ---"

  # Optional fields with their default values
  endpoint = "https://api.cdn77.com"
  timeout  = 10 # in seconds
}
