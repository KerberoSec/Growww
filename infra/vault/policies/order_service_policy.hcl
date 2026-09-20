# Least-privilege Vault policy for Order Service
path "secret/data/growww/orders/*" {
  capabilities = ["read"]
}

path "database/creds/order-service-role" {
  capabilities = ["read"]
}

path "transit/sign/order-hash-signer" {
  capabilities = ["update"]
}
