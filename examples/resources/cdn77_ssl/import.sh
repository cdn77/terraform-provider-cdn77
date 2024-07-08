$ terraform import cdn77_ssl.example <id>,<private_key>

# <id> must be the ID (UUID) of the SSL certificate
# <privateKey> must be an entire private key (including PEM headers) encoded via base64.
# Example:
$ key=$(base64 --wrap=0 key.pem <<EOL
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHqBB2YZkISl1T5TmmZciLN4cJfJPZ6CDpkLgwTiDVyEoAoGCCqGSM49
AwEHoUQDQgAE+lmT51fh5oPIAvtPOEvDw4Ct2sKCt1kYhASlD5b62pT2UyXPrRWp
ekd7UQCYC8K86F1OFeupn2DCOnyGCyK8mw==
-----END EC PRIVATE KEY-----
EOL
)
$ terraform import cdn77_ssl.example "4cd2378b-dec8-49e2-aa17-bf7561452998,$key"
