package goproxy

import (
	crand "crypto/rand"
	"math/rand"
	"strings"

	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

var (
	defaultRootCAPem = []byte(`
-----BEGIN CERTIFICATE-----
MIIFMDCCBBigAwIBAgIGAXWqowQPMA0GCSqGSIb3DQEBCwUAMIGbMSwwKgYDVQQDDCNDaGFybGVz
IFByb3h5IENBICg5IE5vdiAyMDIwLCB3b3JrKTElMCMGA1UECwwcaHR0cHM6Ly9jaGFybGVzcHJv
eHkuY29tL3NzbDERMA8GA1UECgwIWEs3MiBMdGQxETAPBgNVBAcMCEF1Y2tsYW5kMREwDwYDVQQI
DAhBdWNrbGFuZDELMAkGA1UEBhMCTlowIBcNMDAwMTAxMDAwMDAwWhgPMjA1MDAxMDYwMTMzMzFa
MIGbMSwwKgYDVQQDDCNDaGFybGVzIFByb3h5IENBICg5IE5vdiAyMDIwLCB3b3JrKTElMCMGA1UE
CwwcaHR0cHM6Ly9jaGFybGVzcHJveHkuY29tL3NzbDERMA8GA1UECgwIWEs3MiBMdGQxETAPBgNV
BAcMCEF1Y2tsYW5kMREwDwYDVQQIDAhBdWNrbGFuZDELMAkGA1UEBhMCTlowggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQCL3g5heCat/xw11cOvFHuIZfTcaNfRqVDFVRTO4Als+Zu8sApt
/HhysB38iQtS7yRTHJiibaF2I4nEDKFXvkEg1/BGhJ31WKeTErwLIpFFCG5BzCvCT56E/eVeEhGv
F7nc0WjA2qW43QowyGRykx5WLgWu50lgUhGb6fj2+OfkgaZFkFy+VmpUscmmcJCL+EXG5vy9r5E9
X0tCp3shQFF6d2X+eBULfgdtTgW1BCzAeg8w/i5aLDJm9lV3W+kHTIzvYVjZ6dKY4xfNBqyvSUw1
AgnIlav/mq5/tURUdm61AFklAbzPicw7vxeuUTFdysu3yjsPD9IuncxSk5TiM5QnAgMBAAGjggF0
MIIBcDAPBgNVHRMBAf8EBTADAQH/MIIBLAYJYIZIAYb4QgENBIIBHROCARlUaGlzIFJvb3QgY2Vy
dGlmaWNhdGUgd2FzIGdlbmVyYXRlZCBieSBDaGFybGVzIFByb3h5IGZvciBTU0wgUHJveHlpbmcu
IElmIHRoaXMgY2VydGlmaWNhdGUgaXMgcGFydCBvZiBhIGNlcnRpZmljYXRlIGNoYWluLCB0aGlz
IG1lYW5zIHRoYXQgeW91J3JlIGJyb3dzaW5nIHRocm91Z2ggQ2hhcmxlcyBQcm94eSB3aXRoIFNT
TCBQcm94eWluZyBlbmFibGVkIGZvciB0aGlzIHdlYnNpdGUuIFBsZWFzZSBzZWUgaHR0cDovL2No
YXJsZXNwcm94eS5jb20vc3NsIGZvciBtb3JlIGluZm9ybWF0aW9uLjAOBgNVHQ8BAf8EBAMCAgQw
HQYDVR0OBBYEFO142k5RUX2TJg1S3q7zy6jE7rz4MA0GCSqGSIb3DQEBCwUAA4IBAQBxMr2Fypjh
PNzTvYGs7Psz3UV+bxVdL9+H2MzKjOoiCCFN8kH7/lmYiOstH7e4A4NGh4cOBl8AN++j3j1HIe/j
oxLQF8IiPemCS6v1rFrMCo6gulwXWIetOwphCWiyrNHAecLmdgznZENl0HmpAG3TsfD98aiBojrG
VHgsFeUazevPMfwvS8/bpeymR/2D0NEDJt0tzSersxT9kgXcwSLHkb5qx/lwRx9wIa5UCGukbDzS
1asvRUFquL+0RuS/A3Fh/zS87Lf6j9HLE/x0uWlME/Lu+wo9idsug7ytaXDrm30UE6+dWdQ3n0eP
VLNcHV+ZrIqwka94M/t8HavZpm4y
-----END CERTIFICATE-----`)

	defaultRootKeyPem = []byte(`
-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCL3g5heCat/xw1
1cOvFHuIZfTcaNfRqVDFVRTO4Als+Zu8sApt/HhysB38iQtS7yRTHJiibaF2I4nE
DKFXvkEg1/BGhJ31WKeTErwLIpFFCG5BzCvCT56E/eVeEhGvF7nc0WjA2qW43Qow
yGRykx5WLgWu50lgUhGb6fj2+OfkgaZFkFy+VmpUscmmcJCL+EXG5vy9r5E9X0tC
p3shQFF6d2X+eBULfgdtTgW1BCzAeg8w/i5aLDJm9lV3W+kHTIzvYVjZ6dKY4xfN
BqyvSUw1AgnIlav/mq5/tURUdm61AFklAbzPicw7vxeuUTFdysu3yjsPD9IuncxS
k5TiM5QnAgMBAAECggEAfzJiQaHTC0mFji/o1b/62ABgvlFadAFWwx6s3bZA3Cnj
x6UQ1xVTNHmVy6OV+MYicrL+3Dh3LckD8JbL48/RytyWVoskW4tUWhwfhsDY+76/
Bnd3FC70Kl37yaEFFAavsHGAomI1c/kQ22xibQ/99sHDyVvtDvuTqAcPswqJWPRv
odmOhmP3UFpVrR4d9RK4VzJwlA5e52NKxnrjiKreQozNRzI/QAKApgbVFrGWMjmb
2I8YIXf+wzCjip8g4gWwcVHvRJj2Ez8VKY2bzP5l1B9efcCNouuBaTMj0NQtqsXg
kteDju7stb7PzRcdKtJ7GqnrMhZTHxizgvUkOe5sgQKBgQDoTksYlKJ7je6J2P/U
Yw82zdM75kYeEyP/PIYfMYvZdxU7QsMbjZrB7hclG4XX1LoOqgCyU+E/WJZWr2Je
uAoPJbydIzAzFCnkg8xaGbf0e6ITuxQotZOoHgRqUInYq70gP42haAuHaeleVVqs
ztlh+MEStRRaaZo6q2UnWKYjuwKBgQCaIh2KrMAbXDtsTd7sDByc/A4EBsSQ94DW
UzcaFSq26FO8Rnbja9/VGF6wo+k2Umz9oa1ZLb2DQvO5huGBggTnQAvRZxs8OjQc
6dNrJGciLTj7ob/dt1RSXv2y5NOCCrDKJKVJWn5bNDMHTJAS4RIZw0ZvXITB+Bgs
2jTtBp7MhQKBgEDL1dZ9XvTnmemJRZKQLuYycwD6MgShgiDnWOHKiB+YP6vP62v8
C3acWohXLPYOt/bvJFKZYvKwWv7C3MVewC+JbxrFfeRBc43x1UYsdksTURn/zJeu
TglOlhyxakGtZYthLrgetViICjftxuT8rVXOdMwrBgpR+lrzA7v91hmRAoGAPLOG
0uBp39yZAnRAgNHcSu7xTiCkNTtkIAQxxTHk2pfwsktF8xa+1ht83zAOXnhjuBd+
P4rGAfXSKpS2JtzftXsBrHxgu31onKJxwtZZT5pjwKXY/CaBLNeALn3z1lkDevin
p5XeAWkzV4KNkwHUsRS4no7fMczVKITfJyHeVEkCgYBm+bEdoUVSXsNU0IjNGKlr
AcBM4wjTqiqowfVTafsSXOo6600EQDeDfGnQCQLswyIYgiyYUOuRtYv+9YyqnExv
W5hAPG2wpLYEzD6587OQyyREA7j+eljE3QXZdnYmVu2zOur/PBDy0pwfL/An1ooe
nspNHPwqj+F0Uf2zKEUsXA==
-----END PRIVATE KEY-----`)
)

type Cache interface {
	Set(host string, c *tls.Certificate)
	Get(host string) *tls.Certificate
}

var (
	defaultRootCA  *x509.Certificate
	defaultRootKey *rsa.PrivateKey
)

func init() {
	var err error
	block, _ := pem.Decode(defaultRootCAPem)
	defaultRootCA, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(fmt.Errorf("load root cert: %s", err))
	}
	block, _ = pem.Decode(defaultRootKeyPem)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic(fmt.Errorf("load root key: %s", err))
	}

	defaultRootKey = key.(*rsa.PrivateKey)
}

type Certificate struct {
	cache Cache
}

type Pair struct {
	Cert            *x509.Certificate
	CertBytes       []byte
	PrivateKey      *rsa.PrivateKey
	PrivateKeyBytes []byte
}

func NewCertificate(cache Cache) *Certificate {
	return &Certificate{
		cache: cache,
	}
}

// GenerateTlsConfig, get tls config by host
func (c *Certificate) GenerateTlsConfig(host string) (*tls.Config, error) {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	// cache
	if c.cache != nil {
		if cert := c.cache.Get(host); cert != nil {
			tlsConf := &tls.Config{
				Certificates: []tls.Certificate{*cert},
			}
			return tlsConf, nil
		}
	}

	// generate
	pair, err := c.Pair(host, 1, defaultRootCA, defaultRootKey)
	if err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair(pair.CertBytes, pair.PrivateKeyBytes)
	if err != nil {
		return nil, err
	}
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	if c.cache != nil {
		c.cache.Set(host, &cert)
	}

	return tlsConf, nil
}

func (c *Certificate) Pair(host string, expireDays int, rootCA *x509.Certificate, rootKey *rsa.PrivateKey) (*Pair, error) {
	// private key
	priv, err := rsa.GenerateKey(crand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// cert
	cert := c.template(host, expireDays)
	derBytes, err := x509.CreateCertificate(crand.Reader, cert, rootCA, &priv.PublicKey, rootKey)
	if err != nil {
		return nil, err
	}
	serverCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}
	serverKey := pem.EncodeToMemory(keyBlock)

	p := &Pair{
		Cert:            cert,
		CertBytes:       serverCert,
		PrivateKey:      priv,
		PrivateKeyBytes: serverKey,
	}

	return p, nil
}

func (c *Certificate) template(host string, expireYears int) *x509.Certificate {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: host,
		},
		NotBefore:             time.Now().AddDate(-1, 0, 0),
		NotAfter:              time.Now().AddDate(expireYears, 0, 0),
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment,
		EmailAddresses:        []string{"qingqianludao@gmail.com"},
	}
	hosts := strings.Split(host, ",")
	for _, item := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			cert.IPAddresses = append(cert.IPAddresses, ip)
		} else {
			cert.DNSNames = append(cert.DNSNames, item)
		}
	}

	return cert
}

func (c *Certificate) GenerateCA() (*Pair, error) {
	priv, err := rsa.GenerateKey(crand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(rand.Int63()),
		Subject: pkix.Name{
			CommonName:   "Mars",
			Country:      []string{"China"},
			Organization: []string{"4399.com"},
			Province:     []string{"FuJian"},
			Locality:     []string{"Xiamen"},
		},
		NotBefore:             time.Now().AddDate(0, -1, 0),
		NotAfter:              time.Now().AddDate(30, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		EmailAddresses:        []string{"qingqianludao@gmail.com"},
	}

	derBytes, err := x509.CreateCertificate(crand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}
	serverCert := pem.EncodeToMemory(certBlock)

	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}
	serverKey := pem.EncodeToMemory(keyBlock)

	p := &Pair{
		Cert:            tmpl,
		CertBytes:       serverCert,
		PrivateKey:      priv,
		PrivateKeyBytes: serverKey,
	}

	return p, nil
}

func DefaultRootCAPem() []byte {
	return defaultRootCAPem
}
