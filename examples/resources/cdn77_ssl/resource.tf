resource "cdn77_ssl" "example" {
  certificate = file("${path.module}/my-cert.pem")
  private_key = file("${path.module}/my-key.pem")
}
