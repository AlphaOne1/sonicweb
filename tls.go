// SPDX-FileCopyrightText: 2025 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

var errTLSConfig = errors.New("invalid tls configuration")

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

// validateTLSParams validates the provided TLS configuration parameters according to specific constraints.
// Ensures cert and key are both set or unset, acmeDomains mutually exclude cert/key, and clientCAs require TLS setup.
// Returns an error if parameters are invalid, otherwise nil.
func validateTLSParams(cert, key string, acmeDomains, clientCAs []string) error {
	if (len(cert) > 0) != (len(key) > 0) {
		return fmt.Errorf("cert and key must both be given or not given: %w", errTLSConfig)
	}
	if len(cert) > 0 && len(acmeDomains) > 0 {
		return fmt.Errorf("either cert+key or acmeDomains are to be given: %w", errTLSConfig)
	}
	if len(cert) == 0 && len(acmeDomains) == 0 && len(clientCAs) > 0 {
		return fmt.Errorf("clientCAs are only valid if cert+key or acmeDomains are given: %w", errTLSConfig)
	}

	return nil
}

// createCertificateConfig loads a TLS certificate and private key and returns a configured TLS configuration.
// certFile is the path to the certificate file. keyFile is the path to the private key file.
// Returns a tls.Config instance on success or an error if loading the certificate or key fails.
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

// createACMEConfig initializes and returns a TLS configuration for handling ACME-based certificate management.
// acmeDomains specifies the list of allowed domains for certificate provisioning.
// certCache defines the file path where certificates are cached.
// acmeEndpoint is the optional URL for a custom ACME directory endpoint. If empty, the default endpoint is used.
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

// configureClientCAs sets up the ClientCA pool in the provided tls.Config using the
// provided list of CA file paths. It reads each file, appends its certificates to a
// new cert pool, and configures the config for client certificate auth.
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
