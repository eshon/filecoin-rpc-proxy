package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/protofire/filecoin-rpc-proxy/internal/cache"

	"github.com/protofire/filecoin-rpc-proxy/internal/config"
	"github.com/protofire/filecoin-rpc-proxy/internal/logger"

	"github.com/sirupsen/logrus"
)

type Server struct {
	target *url.URL
	logger *logrus.Entry
	proxy  *httputil.ReverseProxy
	transport
}

func NewServer(c *config.Config) (*Server, error) {
	proxyURL, err := url.Parse(c.ProxyURL)
	if err != nil {
		return nil, err
	}
	log := logger.InitLogger(c.LogLevel, c.LogPrettyPrint)
	s := &Server{
		target: proxyURL,
		logger: log,
		proxy:  httputil.NewSingleHostReverseProxy(proxyURL),
		transport: transport{
			logger:  log,
			cache:   cache.NewMemoryCacheFromConfig(c),
			matcher: NewMatcherFromConfig(c),
		},
	}
	s.proxy.Transport = &s.transport
	return s, nil
}

func (p *Server) RPCProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-rpc-proxy", "rpc-proxy")
	p.proxy.ServeHTTP(w, r)
}
