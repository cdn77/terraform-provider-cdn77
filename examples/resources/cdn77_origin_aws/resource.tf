resource "cdn77_origin_aws" "example" {
  label             = "Assets AWS bucket for example.com"
  url               = "https://examplecom-static-assets.s3.eu-central-1.amazonaws.com"
  region            = "eu-central-1"
  access_key_id     = "23478207027842073230762374023"
  access_key_secret = "VWK92izmd7zpY8Khs/Dllv4yLYc4sFWNyg2XtuNF"
}
