# tls.Config 参数解析

> tls.Config 数据结构

```cgo
// Config结构用于配置TLS客户端或服务器.
// 在一个连接通过TLS函数后, 不得对其进行修改. Config可以重用; tls包也不会对其进行修改.
type Config struct {
    // Rand的作用是提供随机数
    // 如果Rand为nil, 则TLS使用软件包 `crypto/rand` 中的加密随机读取器.
    // Reader 必须可以安全地被多个goroutine使用.
	Rand io.Reader
    
    // Time 返回当前时间, 以自该纪元以来的秒数为单位. 如果Time为nil, 则TLS使用time.Now.
	Time func() time.Time
    
    // Certificates包含一个或多个证书链, 以呈现给连接的另一端. 
    // [Server]服务器配置必须至少包含一个证书或设置GetCertificate参数.
    // [Client]进行客户端身份验证的客户端可以设置 `Certificates` 或 `GetClientCertificate`.
	Certificates []Certificate
    
    // NameToCertificate 将证书名称映射到 Certificates. 请注意, 证书名称的格式可以为 "*.example.com", 
    // 也可以不必一定是域名.
    // 请参阅 Config.BuildNameToCertificate
    // 如果值是nil, 则使用 Certificates 的第一个值应用于所有的连接.
	NameToCertificate map[string]*Certificate
    
    // GetCertificate返回基于给定ClientHelloInfo 的 Certificate. 
    // 仅当客户端提供SNI信息或 Certificates 为空时, 才会调用该方法.
    //
    // 如果GetCertificate为nil或返回nil, 则从NameToCertificate检索证书.
    // 如果NameToCertificate为nil, 则将使用 Certificates 的第一个值.
	GetCertificate func(*ClientHelloInfo) (*Certificate, error)
    
    // 如果服务器从客户端请求证书, 则调用GetClientCertificate(如果不为nil). 
    // 如果设置, 则 Certificates 的值将被忽略.
    //
    // 如果GetClientCertificate 返回错误, 则 handshake 将被中止并返回该错误. 否则, 
    // GetClientCertificate必须返回 non-nil Certificate. 
    // 如果Certificate.Certificate为空, 则不会将任何证书发送到服务器. 如果服务器 unacceptable, 则它可能会中止
    // 握手.
    //
    // 如果发生重新协商或正在使用TLS 1.3, 则可为同一连接多次调用GetClientCertificate.
	GetClientCertificate func(*CertificateRequestInfo) (*Certificate, error)

    // 从客户端收到ClientHello之后, 将调用GetConfigForClient(如果不是nil). 
    // 它可能会返回非nil Config, 以更改将用于处理此连接的Config. 
    // 如果返回的Config为nil, 则将使用原始Config. 此回调返回的Config可能随后无法修改.
    //
    // 如果GetConfigForClient为nil, 则传递给 Server() 的Config将用于所有连接.
    //
    // 返回的Config必须是唯一的, 如果未设置, session ticket keys 将从原始Config复制.
    // 具体来说, 如果在原始配置上调用了SetSessionTicketKeys而不是在返回的配置上调用了SetSessionTicketKeys,
    // 则在使用之前, 原始配置中的 ticket keys 将被复制到新配置中. 否则, 如果在原始配置中设置了 SessionTicketKey
    // 而不在返回的配置中设置了SessionTicketKey, 则它将在使用前复制到返回的配置中. 
    // 如果这两种情况均不适用, 则返回的配置中的密钥材料将用于会话票证.

	// GetConfigForClient, if not nil, is called after a ClientHello is
	// received from a client. It may return a non-nil Config in order to
	// change the Config that will be used to handle this connection. If
	// the returned Config is nil, the original Config will be used. The
	// Config returned by this callback may not be subsequently modified.
	//
	// If GetConfigForClient is nil, the Config passed to Server() will be
	// used for all connections.
	//
	// Uniquely for the fields in the returned Config, session ticket keys
	// will be duplicated from the original Config if not set.
	// Specifically, if SetSessionTicketKeys was called on the original
	// config but not on the returned config then the ticket keys from the
	// original config will be copied into the new config before use.
	// Otherwise, if SessionTicketKey was set in the original config but
	// not in the returned config then it will be copied into the returned
	// config before use. If neither of those cases applies then the key
	// material from the returned config will be used for session tickets.
	GetConfigForClient func(*ClientHelloInfo) (*Config, error)

	// VerifyPeerCertificate, if not nil, is called after normal
	// certificate verification by either a TLS client or server. It
	// receives the raw ASN.1 certificates provided by the peer and also
	// any verified chains that normal processing found. If it returns a
	// non-nil error, the handshake is aborted and that error results.
	//
	// If normal verification fails then the handshake will abort before
	// considering this callback. If normal verification is disabled by
	// setting InsecureSkipVerify, or (for a server) when ClientAuth is
	// RequestClientCert or RequireAnyClientCert, then this callback will
	// be considered but the verifiedChains argument will always be nil.
	VerifyPeerCertificate func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error

	// RootCAs defines the set of root certificate authorities
	// that clients use when verifying server certificates.
	// If RootCAs is nil, TLS uses the host's root CA set.
	RootCAs *x509.CertPool

	// NextProtos is a list of supported application level protocols, in
	// order of preference.
	NextProtos []string

	// ServerName is used to verify the hostname on the returned
	// certificates unless InsecureSkipVerify is given. It is also included
	// in the client's handshake to support virtual hosting unless it is
	// an IP address.
	ServerName string

	// ClientAuth determines the server's policy for
	// TLS Client Authentication. The default is NoClientCert.
	ClientAuth ClientAuthType

	// ClientCAs defines the set of root certificate authorities
	// that servers use if required to verify a client certificate
	// by the policy in ClientAuth.
	ClientCAs *x509.CertPool

	// InsecureSkipVerify controls whether a client verifies the
	// server's certificate chain and host name.
	// If InsecureSkipVerify is true, TLS accepts any certificate
	// presented by the server and any host name in that certificate.
	// In this mode, TLS is susceptible to man-in-the-middle attacks.
	// This should be used only for testing.
	InsecureSkipVerify bool

	// CipherSuites is a list of supported cipher suites for TLS versions up to
	// TLS 1.2. If CipherSuites is nil, a default list of secure cipher suites
	// is used, with a preference order based on hardware performance. The
	// default cipher suites might change over Go versions. Note that TLS 1.3
	// ciphersuites are not configurable.
	CipherSuites []uint16

	// PreferServerCipherSuites controls whether the server selects the
	// client's most preferred ciphersuite, or the server's most preferred
	// ciphersuite. If true then the server's preference, as expressed in
	// the order of elements in CipherSuites, is used.
	PreferServerCipherSuites bool

	// SessionTicketsDisabled may be set to true to disable session ticket and
	// PSK (resumption) support. Note that on clients, session ticket support is
	// also disabled if ClientSessionCache is nil.
	SessionTicketsDisabled bool

	// SessionTicketKey is used by TLS servers to provide session resumption.
	// See RFC 5077 and the PSK mode of RFC 8446. If zero, it will be filled
	// with random data before the first server handshake.
	//
	// If multiple servers are terminating connections for the same host
	// they should all have the same SessionTicketKey. If the
	// SessionTicketKey leaks, previously recorded and future TLS
	// connections using that key might be compromised.
	SessionTicketKey [32]byte

	// ClientSessionCache is a cache of ClientSessionState entries for TLS
	// session resumption. It is only used by clients.
	ClientSessionCache ClientSessionCache

	// MinVersion contains the minimum SSL/TLS version that is acceptable.
	// If zero, then TLS 1.0 is taken as the minimum.
	MinVersion uint16

	// MaxVersion contains the maximum SSL/TLS version that is acceptable.
	// If zero, then the maximum version supported by this package is used,
	// which is currently TLS 1.3.
	MaxVersion uint16

	// CurvePreferences contains the elliptic curves that will be used in
	// an ECDHE handshake, in preference order. If empty, the default will
	// be used. The client will use the first preference as the type for
	// its key share in TLS 1.3. This may change in the future.
	CurvePreferences []CurveID

	// DynamicRecordSizingDisabled disables adaptive sizing of TLS records.
	// When true, the largest possible TLS record size is always used. When
	// false, the size of TLS records may be adjusted in an attempt to
	// improve latency.
	DynamicRecordSizingDisabled bool

	// Renegotiation controls what types of renegotiation are supported.
	// The default, none, is correct for the vast majority of applications.
	Renegotiation RenegotiationSupport

	// KeyLogWriter optionally specifies a destination for TLS master secrets
	// in NSS key log format that can be used to allow external programs
	// such as Wireshark to decrypt TLS connections.
	// See https://developer.mozilla.org/en-US/docs/Mozilla/Projects/NSS/Key_Log_Format.
	// Use of KeyLogWriter compromises security and should only be
	// used for debugging.
	KeyLogWriter io.Writer

	serverInitOnce sync.Once // guards calling (*Config).serverInit

	// mutex protects sessionTicketKeys.
	mutex sync.RWMutex
	// sessionTicketKeys contains zero or more ticket keys. If the length
	// is zero, SessionTicketsDisabled must be true. The first key is used
	// for new tickets and any subsequent keys can be used to decrypt old
	// tickets.
	sessionTicketKeys []ticketKey
}
```