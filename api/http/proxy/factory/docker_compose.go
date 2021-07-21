package factory

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
	"github.com/portainer/portainer/api/http/proxy/factory/dockercompose"
)

// ProxyServer provide an extedned proxy with a local server to forward requests
type ProxyServer struct {
	server *http.Server
	Port   int
}

func (factory *ProxyFactory) NewDockerComposeAgentProxy(endpoint *portainer.Endpoint) (*ProxyServer, error) {

	if endpoint.Type == portainer.EdgeAgentOnDockerEnvironment {
		return &ProxyServer{
			Port: factory.reverseTunnelService.GetTunnelDetails(endpoint.ID).Port,
		}, nil
	}

	endpointURL, err := url.Parse(endpoint.URL)
	if err != nil {
		return nil, err
	}

	endpointURL.Scheme = "http"
	httpTransport := &http.Transport{}

	if endpoint.TLSConfig.TLS || endpoint.TLSConfig.TLSSkipVerify {
		config, err := crypto.CreateTLSConfigurationFromDisk(endpoint.TLSConfig.TLSCACertPath, endpoint.TLSConfig.TLSCertPath, endpoint.TLSConfig.TLSKeyPath, endpoint.TLSConfig.TLSSkipVerify)
		if err != nil {
			return nil, err
		}

		httpTransport.TLSClientConfig = config
		endpointURL.Scheme = "https"
	}

	proxy := newSingleHostReverseProxyWithHostHeader(endpointURL)

	proxy.Transport = dockercompose.NewAgentTransport(factory.signatureService, httpTransport)

	proxyServer := &ProxyServer{
		server: &http.Server{
			Handler: proxy,
		},
		Port: 0,
	}

	err = proxyServer.start()
	if err != nil {
		return nil, err
	}

	return proxyServer, nil
}

func (proxy *ProxyServer) start() error {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	proxy.Port = listener.Addr().(*net.TCPAddr).Port
	go func() {
		proxyHost := fmt.Sprintf("127.0.0.1:%d", proxy.Port)
		log.Printf("Starting Proxy server on %s...\n", proxyHost)

		err := proxy.server.Serve(listener)
		log.Printf("Exiting Proxy server %s\n", proxyHost)

		if err != nil && err != http.ErrServerClosed {
			log.Printf("Proxy server %s exited with an error: %s\n", proxyHost, err)
		}
	}()

	return nil
}

// Close shuts down the server
func (proxy *ProxyServer) Close() {
	if proxy.server != nil {
		proxy.server.Close()
	}
}
