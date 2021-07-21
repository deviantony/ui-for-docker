package websocket

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	"github.com/portainer/portainer/api"
)

func (handler *Handler) proxyEdgeAgentWebsocketRequest(w http.ResponseWriter, r *http.Request, params *webSocketRequestParams) error {
	tunnel := handler.ReverseTunnelService.GetTunnelDetails(params.endpoint.ID)

	endpointURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", tunnel.Port))
	if err != nil {
		return err
	}

	endpointURL.Scheme = "ws"
	proxy := websocketproxy.NewProxy(endpointURL)

	proxy.Director = func(incoming *http.Request, out http.Header) {
		out.Set(portainer.PortainerAgentTargetHeader, params.nodeName)
		out.Set(portainer.PortainerAgentKubernetesSATokenHeader, params.token)
	}

	handler.ReverseTunnelService.SetTunnelStatusToActive(params.endpoint.ID)
	proxy.ServeHTTP(w, r)

	return nil
}

func (handler *Handler) proxyAgentWebsocketRequest(w http.ResponseWriter, r *http.Request, params *webSocketRequestParams) error {
	endpointURL := params.endpoint.URL
	if params.endpoint.Type == portainer.AgentOnKubernetesEnvironment {
		endpointURL = fmt.Sprintf("http://%s", params.endpoint.URL)
	}

	agentURL, err := url.Parse(endpointURL)
	if err != nil {
		return err
	}

	agentURL.Scheme = "ws"
	proxy := websocketproxy.NewProxy(agentURL)

	if params.endpoint.TLSConfig.TLS || params.endpoint.TLSConfig.TLSSkipVerify {
		agentURL.Scheme = "wss"
		proxy.Dialer = &websocket.Dialer{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: params.endpoint.TLSConfig.TLSSkipVerify,
			},
		}
	}

	signature, err := handler.SignatureService.CreateSignature(portainer.PortainerAgentSignatureMessage)
	if err != nil {
		return err
	}

	proxy.Director = func(incoming *http.Request, out http.Header) {
		out.Set(portainer.PortainerAgentPublicKeyHeader, handler.SignatureService.EncodedPublicKey())
		out.Set(portainer.PortainerAgentSignatureHeader, signature)
		out.Set(portainer.PortainerAgentTargetHeader, params.nodeName)
		out.Set(portainer.PortainerAgentKubernetesSATokenHeader, params.token)
	}

	proxy.ServeHTTP(w, r)

	return nil
}
