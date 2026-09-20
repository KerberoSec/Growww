package growww.authz

default allow = false

# Allow public health and status endpoints
allow {
    input.path = "/healthz"
}

allow {
    input.path = "/api/v1/public/market-data/ticker"
}

# Protected routes require valid JWT token with non-expired timestamp
token_valid {
    input.token.exp > time.now_ns() / 1000000000
    not token_blacklisted
}

token_blacklisted {
    input.blacklisted_jtis[input.token.jti]
}

# Trading orders require TIER_2 KYC and TRADER role
allow {
    token_valid
    startswith(input.path, "/api/v1/orders")
    input.token.kyc_tier == "TIER_2"
    input.token.roles[_] == "TRADER"
    input.token.mfa_verified == true
    input.token.entity == input.requested_entity
}

# Compliance backoffice routes require COMPLIANCE_OFFICER role and MFA
allow {
    token_valid
    startswith(input.path, "/api/v1/compliance")
    input.token.roles[_] == "COMPLIANCE_OFFICER"
    input.token.mfa_verified == true
}

# Admin superuser operations
allow {
    token_valid
    startswith(input.path, "/api/v1/admin")
    input.token.roles[_] == "SUPER_ADMIN"
    input.token.mfa_verified == true
}
