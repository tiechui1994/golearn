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
    // 对于返回的Config中的字段, 唯一的是, 如果未设置, 会话票据密钥(session ticket key)将从原始Config复制.
    // 具体来说, 如果在原始 Config 上调用了SetSessionTicketKeys而不是在返回的 Config 上调用了SetSessionTicketKeys,
    // 则原始配置中的票证密钥将在使用前复制到新配置中.
    // 否则, 如果SessionTicketKey是在原始配置中设置的, 而不是在返回的配置中设置的, 则它将在使用前复制到返回的配置中.
    // 如果这两种情况均不适用, 则返回的配置中的密钥材料将用于会话票证.
	GetConfigForClient func(*ClientHelloInfo) (*Config, error)

    // TLS客户端或服务器在常规证书验证后调用VerifyPeerCertificate(如果不为nil). 
    // 它接收对等方(peer)提供的原始ASN.1证书以及正常处理发现的所有已验证链. 
    // 如果返回非零错误, 则握手(handshake)中止, 并导致该错误.
    //
    // 如果正常验证失败, 则握手将中止. 然后再考虑此回调(正常验证成功的前提下, 参考 doFullHandshake 等函数的逻辑).
    // 如果通过设置InsecureSkipVerify禁用了正常验证, 或者(对于server而言) 当ClientAuth为
    // RequestClientCert或RequireAnyClientCert时, 则将考虑此回调, 但verifiedChains参数始终为nil.
	VerifyPeerCertificate func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error

    // RootCA定义了客户端在验证服务器证书时使用的一组根证书集合.
    // 如果RootCAs为nil, 则TLS使用主机的根CA集合.
	RootCAs *x509.CertPool

    // NextProtos按优先顺序列出了受支持的应用程序级别协议.
	NextProtos []string

    // 除非设置InsecureSkipVerify, 否则ServerName用于验证返回的证书上的主机名.
    // 除非它是IP地址, 否则它也包含在客户端的握手中以支持虚拟主机.
	ServerName string

    // ClientAuth 确定 TLS Client Authentication 服务器的策略. 默认值为NoClientCert.
	ClientAuth ClientAuthType
    
    // ClientCAs定义了一组根证书集合, 如果ClientAuth中的策略要求验证客户端证书, 服务器将使用这些证书.
	ClientCAs *x509.CertPool
    
    // InsecureSkipVerify控制客户端是否验证服务器的证书链和主机名.
    // 如果InsecureSkipVerify为true, 则TLS接受服务器提供的任何证书以及该证书中的任何主机名.
    // 在这种模式下, TLS容易受到中间人攻击。. 这仅应用于测试.
	InsecureSkipVerify bool
    
    // CipherSuites是受支持的密码套件的列表, 适用于TLS 1.2以下的TLS版本. 
    // 如果CipherSuites为nil, 则使用默认的安全密码套件列表, 其优先级顺序基于硬件性能. 
    // 默认密码套件可能会在Go版本上更改. 请注意, TLS 1.3密码套件不可配置.
	CipherSuites []uint16

    // PreferServerCipherSuites控制服务器是选择客户端的首选密码套件, 还是选择服务器的首选密码套件.
    // 如果为true, 则使用按CipherSuites中的元素顺序表示的服务器首选项.
	PreferServerCipherSuites bool
    
    // SessionTicketsDisabled可以设置为true以禁用会话票证和PSK(resumption [恢复])支持. 
    // 请注意, 在客户端上, 如果ClientSessionCache为nil, 会话票证支持也会被禁用.
	SessionTicketsDisabled bool
    
    // [session resumption] https://wiki.jikexueyuan.com/project/openresty/ssl/session_resumption.html
    //
    // TLS服务器使用SessionTicketKey提供会话恢复(session resumption).
    // 请参阅RFC 5077和RFC 8446的PSK模式. 如果为nil, 它将在第一次服务器握手之前填充随机数据.
    //
    // 如果多个服务器正在终止同一主机的连接, 则它们都应具有相同的 SessionTicketKey。.
    // 如果SessionTicketKey泄漏, 则先前使用该密钥记录的TLS连接和将来的TLS连接可能会受到影响.
	SessionTicketKey [32]byte

	// ClientSessionCache is a cache of ClientSessionState entries for TLS
	// session resumption. It is only used by clients.
	ClientSessionCache ClientSessionCache

	// MinVersion包含可接受的最低SSL/TLS版本. 如果为零, 则TLS 1.0为最小值.
	MinVersion uint16

	// MaxVersion包含可接受的最大SSL/TLS版本.
    // 如果为0, 则使用此程序包支持的最高版本, 当前版本为TLS 1.3.
	MaxVersion uint16
    
    // CurvePreferences包含将以优先顺序在ECDHE握手中使用的椭圆曲线. 
    // 如果为空, 将使用默认值. 客户端将使用第一个首选项作为其TLS 1.3中密钥共享的类型. 
    // 将来可能会改变.
	CurvePreferences []CurveID
    
    // DynamicRecordSizingDisabled 是否禁用TLS记录的自适应大小调整.
    // 如果为true, 则始终使用最大的TLS记录大小, 如果为false, 则可以调整TLS记录的大小, 以尝试改善延迟.
	DynamicRecordSizingDisabled bool
    
    // Renegotiation控制支持什么类型的重新协商(renegotiation).
    // 对于大多数应用程序, 默认值none是正确的.
	Renegotiation RenegotiationSupport
    
    // KeyLogWriter可以选择以NSS密钥日志格式指定TLS主密钥的目的地, 该目的地可用于允许Wireshark等外部程序
    // 解密TLS连接.
    // 参见https://developer.mozilla.org/enUS/docs/Mozilla/Projects/NSS/Key_Log_Format.
    // 使用KeyLogWriter会损害安全性, 应仅用于调试.
	KeyLogWriter io.Writer

	serverInitOnce sync.Once // guards calling (*Config).serverInit

	// mutex protects sessionTicketKeys.
	mutex sync.RWMutex
	
	// sessionTicketKeys包含零个或多个票证密钥.
	// 如果长度为零, 则SessionTicketsDisabled必须为true.
	// 第一个密钥用于新票据, 而任何后续密钥均可用于解密旧票据.
	sessionTicketKeys []ticketKey
}
```