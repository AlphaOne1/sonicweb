// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

// generateTLSConfig generates a new TLS configuration if the parameters are set accordingly.
// To use a user-supplied cert- and key file, only specify those two parameters. Specifying
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

	if err := validateTLSParams(cert, key, acmeDomains, clientCAs); err != nil {
		return nil, err
	}

	// completely valid, we do not have a TLS configuration
	if len(cert) == 0 && len(acmeDomains) == 0 {
		return nil, nil
	}

	var config *tls.Config
	var err error

	if len(cert) > 0 {
		config, err = createCertificateConfig(cert, key)

		if err != nil {
			return nil, err
		}
	}

	if len(acmeDomains) > 0 {
		config = createACMEConfig(acmeDomains, certCache, acmeEndpoint)
	}

	if config != nil && len(clientCAs) > 0 {
		if err := configureClientCAs(config, clientCAs); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func validateTLSParams(cert, key string, acmeDomains, clientCAs []string) error {
	if (len(cert) > 0) != (len(key) > 0) {
		return fmt.Errorf("invalid tls config, cert and key must both be given or not given")
	}
	if len(cert) > 0 && len(acmeDomains) > 0 {
		return fmt.Errorf("either cert+key or acmeDomains are to be given")
	}
	if len(cert) == 0 && len(acmeDomains) == 0 && len(clientCAs) > 0 {
		return fmt.Errorf("clientCAs are only valid if cert+key or acmeDomains are given")
	}

	return nil
}

func createCertificateConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		return nil, fmt.Errorf("could not load certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}, nil
}

func createACMEConfig(acmeDomains []string, certCache, acmeEndpoint string) *tls.Config {
	var acmeClient *acme.Client

	if len(acmeEndpoint) > 0 {
		acmeClient = &acme.Client{
			DirectoryURL: acmeEndpoint,
		}
	}

	certManager := autocert.Manager{
		Cache:      autocert.DirCache(certCache),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(acmeDomains...),
		Client:     acmeClient,
	}

	return certManager.TLSConfig()
}

func configureClientCAs(config *tls.Config, clientCAs []string) error {
	clientCAPool := x509.NewCertPool()

	for _, ca := range clientCAs {
		caFile, err := os.ReadFile(filepath.Clean(ca))

		if err != nil {
			return fmt.Errorf("could not read client CA file: %w", err)
		}

		clientCAPool.AppendCertsFromPEM(caFile)
	}

	config.ClientCAs = clientCAPool
	config.ClientAuth = tls.RequireAndVerifyClientCert

	return nil
}
