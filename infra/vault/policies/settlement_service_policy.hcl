# Least-privilege Vault policy for Settlement Service
path "secret/data/growww/settlement/*" {
  capabilities = ["read"]
}

path "database/creds/settlement-service-role" {
  capabilities = ["read"]
}

path "transit/sign/growww-besu-relayer" {
  capabilities = ["update"]
}
