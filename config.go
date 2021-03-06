package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"

	"code.cloudfoundry.org/go-envstruct"
	"google.golang.org/grpc/credentials"
)

// Config is the configuration for a LogCache.
type Config struct {
	Addr string `env:"ADDR, required, report"`
	APIServer string `env:"API_SERVER, required, report"`
	Token string `env:"TOKEN, required, report"`
	AppSelector string `env:"APP_SELECTOR, required, report"`
	Namespace string `env:"NAMESPACE"`

	// QueryTimeout sets the maximum allowed runtime for a single PromQL query.
	// Smaller timeouts are recommended.
	QueryTimeout int64 `env:"QUERY_TIMEOUT, report"`

	TLS
}

// LoadConfig creates Config object from environment variables
func LoadConfig() (*Config, error) {
	c := Config{
		//Addr:         ":8080",
		QueryTimeout: 10,
	}

	if err := envstruct.Load(&c); err != nil {
		return nil, err
	}

	return &c, nil
}


type TLS struct {
	CAPath   string `env:"CA_PATH,   required, report"`
	CertPath string `env:"CERT_PATH, required, report"`
	KeyPath  string `env:"KEY_PATH,  required, report"`
}

func (t TLS) Credentials(cn string) credentials.TransportCredentials {
	creds, err := NewTLSCredentials(t.CAPath, t.CertPath, t.KeyPath, cn)
	if err != nil {
		log.Fatalf("failed to load TLS config: %s", err)
	}

	return creds
}

func NewTLSCredentials(
	caPath string,
	certPath string,
	keyPath string,
	cn string,
) (credentials.TransportCredentials, error) {
	cfg, err := NewMutualTLSConfig(caPath, certPath, keyPath, cn)
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(cfg), nil
}

func NewMutualTLSConfig(caPath, certPath, keyPath, cn string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	tlsConfig := NewBaseTLSConfig()
	tlsConfig.ServerName = cn
	tlsConfig.Certificates = []tls.Certificate{cert}

	caCertBytes, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertBytes); !ok {
		return nil, errors.New("cannot parse ca cert")
	}

	tlsConfig.RootCAs = caCertPool

	return tlsConfig, nil
}

func NewBaseTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
		CipherSuites:       supportedCipherSuites,
	}
}

var supportedCipherSuites = []uint16{
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
}
