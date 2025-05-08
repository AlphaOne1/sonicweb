// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

// generateTLSConfig generates a new TLS configuration, if the parameters are accordingly.
// To use a use supplied cert- and key file, just specify those two parameters. Specifying
// the acmeDomains will lead to an error in this case.
// To use the Let's Encrypt feature, cert and key are to be left empty and acmeDomains must
// be specified.
// If nothing is specified, no TLS configuration is generated.
func generateTLSConfig(
	cert string,
	key string,
	acmeDomains []string,
	certCache string,
	acmeEndpoint string,
	clientCAs []string) (*tls.Config, error) {
	if (len(cert) > 0) != (len(key) > 0) {
		return nil, fmt.Errorf("invalid tls config, cert and key must both be given or not given")
	}

	if len(cert) > 0 && len(acmeDomains) > 0 {
		return nil, fmt.Errorf("either cert+key or acmeDomains are to be given")
	}

	if len(cert) == 0 && len(acmeDomains) == 0 {
		if len(clientCAs) > 0 {
			return nil, fmt.Errorf("clientCAs are only valid if cert+key or acmeDomains are given")
		}

		// completely valid, we do not have a TLS config
		return nil, nil
	}

	var result *tls.Config

	if len(cert) > 0 {
		cert, err := tls.LoadX509KeyPair(cert, key)

		if err != nil {
			return nil, fmt.Errorf("could not load certificate: %w", err)
		}

		result = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	if len(acmeDomains) > 0 {
		var acmeClient *acme.Client

		if len(acmeEndpoint) > 0 {
			acmeClient = &acme.Client{
				DirectoryURL: acmeEndpoint,
			}
		}

		// automatic certificate management with autocert
		certManager := autocert.Manager{
			Cache:      autocert.DirCache(certCache),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(acmeDomains...),
			Client:     acmeClient,
		}

		result = certManager.TLSConfig()
	}

	if len(clientCAs) > 0 {
		var clientCAPool *x509.CertPool = x509.NewCertPool()

		for _, ca := range clientCAs {
			caFile, caFileErr := os.ReadFile(ca)

			if caFileErr != nil {
				return nil, fmt.Errorf("could not read client CA file: %w", caFileErr)
			}

			clientCAPool.AppendCertsFromPEM(caFile)
		}

		result.ClientCAs = clientCAPool
		result.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return result, nil
}
