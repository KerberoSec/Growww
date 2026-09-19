package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
)

type ZeroTrustMeshConfig struct {
	AllowedSPIFFEIDs map[string]bool
	ServerCert       tls.Certificate
	ClientCAPool     *x509.CertPool
}

func NewZeroTrustMeshConfig() *ZeroTrustMeshConfig {
	return &ZeroTrustMeshConfig{
		AllowedSPIFFEIDs: map[string]bool{
			"spiffe://growww.internal/ns/core/sa/order-service":          true,
			"spiffe://growww.internal/ns/core/sa/matching-engine":        true,
			"spiffe://growww.internal/ns/core/sa/wallet-account-service": true,
			"spiffe://growww.internal/ns/core/sa/trade-settlement":       true,
		},
	}
}

// VerifyClientSPIFFEID validates peer mTLS SAN URI according to SPIFFE zero-trust identity
func (c *ZeroTrustMeshConfig) VerifyClientSPIFFEID(r *http.Request) error {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return errors.New("mTLS required: no client certificate presented")
	}

	cert := r.TLS.PeerCertificates[0]
	if len(cert.URIs) == 0 {
		return errors.New("client certificate missing SPIFFE ID SAN")
	}

	spiffeID := cert.URIs[0].String()
	if !c.AllowedSPIFFEIDs[spiffeID] {
		return errors.New("unauthorized SPIFFE ID identity in service mesh")
	}

	return nil
}
