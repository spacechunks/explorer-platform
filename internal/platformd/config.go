package platformd

type Config struct {
	ProxyServiceListenSock string
	CRIListenSock          string
	EnvoyImage             string
}
