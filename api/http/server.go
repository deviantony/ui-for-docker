package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/adminmonitor"
	"github.com/portainer/portainer/api/crypto"
	"github.com/portainer/portainer/api/docker"
	"github.com/portainer/portainer/api/http/handler"
	"github.com/portainer/portainer/api/http/handler/auth"
	"github.com/portainer/portainer/api/http/handler/backup"
	"github.com/portainer/portainer/api/http/handler/customtemplates"
	"github.com/portainer/portainer/api/http/handler/edgegroups"
	"github.com/portainer/portainer/api/http/handler/edgejobs"
	"github.com/portainer/portainer/api/http/handler/edgestacks"
	"github.com/portainer/portainer/api/http/handler/edgetemplates"
	"github.com/portainer/portainer/api/http/handler/endpointedge"
	"github.com/portainer/portainer/api/http/handler/endpointgroups"
	"github.com/portainer/portainer/api/http/handler/endpointproxy"
	"github.com/portainer/portainer/api/http/handler/endpoints"
	"github.com/portainer/portainer/api/http/handler/file"
	"github.com/portainer/portainer/api/http/handler/motd"
	"github.com/portainer/portainer/api/http/handler/registries"
	"github.com/portainer/portainer/api/http/handler/resourcecontrols"
	"github.com/portainer/portainer/api/http/handler/roles"
	"github.com/portainer/portainer/api/http/handler/settings"
	sslhandler "github.com/portainer/portainer/api/http/handler/ssl"
	"github.com/portainer/portainer/api/http/handler/stacks"
	"github.com/portainer/portainer/api/http/handler/status"
	"github.com/portainer/portainer/api/http/handler/tags"
	"github.com/portainer/portainer/api/http/handler/teammemberships"
	"github.com/portainer/portainer/api/http/handler/teams"
	"github.com/portainer/portainer/api/http/handler/templates"
	"github.com/portainer/portainer/api/http/handler/upload"
	"github.com/portainer/portainer/api/http/handler/users"
	"github.com/portainer/portainer/api/http/handler/webhooks"
	"github.com/portainer/portainer/api/http/handler/websocket"
	"github.com/portainer/portainer/api/http/offlinegate"
	"github.com/portainer/portainer/api/http/proxy"
	"github.com/portainer/portainer/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/authorization"
	"github.com/portainer/portainer/api/internal/ssl"
	"github.com/portainer/portainer/api/kubernetes/cli"
)

// Server implements the portainer.Server interface
type Server struct {
	AuthorizationService        *authorization.Service
	BindAddress                 string
	BindAddressHTTPS            string
	HTTPEnabled                 bool
	AssetsPath                  string
	Status                      *portainer.Status
	ReverseTunnelService        portainer.ReverseTunnelService
	ComposeStackManager         portainer.ComposeStackManager
	CryptoService               portainer.CryptoService
	SignatureService            portainer.DigitalSignatureService
	SnapshotService             portainer.SnapshotService
	FileService                 portainer.FileService
	DataStore                   portainer.DataStore
	GitService                  portainer.GitService
	JWTService                  portainer.JWTService
	LDAPService                 portainer.LDAPService
	OAuthService                portainer.OAuthService
	SwarmStackManager           portainer.SwarmStackManager
	ProxyManager                *proxy.Manager
	KubernetesTokenCacheManager *kubernetes.TokenCacheManager
	Handler                     *handler.Handler
	SSLService                  *ssl.Service
	DockerClientFactory         *docker.ClientFactory
	KubernetesClientFactory     *cli.ClientFactory
	KubernetesDeployer          portainer.KubernetesDeployer
	ShutdownCtx                 context.Context
	ShutdownTrigger             context.CancelFunc
}

// Start starts the HTTP server
func (server *Server) Start() error {
	kubernetesTokenCacheManager := server.KubernetesTokenCacheManager

	requestBouncer := security.NewRequestBouncer(server.DataStore, server.JWTService)

	rateLimiter := security.NewRateLimiter(10, 1*time.Second, 1*time.Hour)
	offlineGate := offlinegate.NewOfflineGate()

	var authHandler = auth.NewHandler(requestBouncer, rateLimiter)
	authHandler.DataStore = server.DataStore
	authHandler.CryptoService = server.CryptoService
	authHandler.JWTService = server.JWTService
	authHandler.LDAPService = server.LDAPService
	authHandler.ProxyManager = server.ProxyManager
	authHandler.KubernetesTokenCacheManager = kubernetesTokenCacheManager
	authHandler.OAuthService = server.OAuthService

	adminMonitor := adminmonitor.New(5*time.Minute, server.DataStore, server.ShutdownCtx)
	adminMonitor.Start()

	var backupHandler = backup.NewHandler(requestBouncer, server.DataStore, offlineGate, server.FileService.GetDatastorePath(), server.ShutdownTrigger, adminMonitor)

	var roleHandler = roles.NewHandler(requestBouncer)
	roleHandler.DataStore = server.DataStore

	var customTemplatesHandler = customtemplates.NewHandler(requestBouncer)
	customTemplatesHandler.DataStore = server.DataStore
	customTemplatesHandler.FileService = server.FileService
	customTemplatesHandler.GitService = server.GitService

	var edgeGroupsHandler = edgegroups.NewHandler(requestBouncer)
	edgeGroupsHandler.DataStore = server.DataStore

	var edgeJobsHandler = edgejobs.NewHandler(requestBouncer)
	edgeJobsHandler.DataStore = server.DataStore
	edgeJobsHandler.FileService = server.FileService
	edgeJobsHandler.ReverseTunnelService = server.ReverseTunnelService

	var edgeStacksHandler = edgestacks.NewHandler(requestBouncer)
	edgeStacksHandler.DataStore = server.DataStore
	edgeStacksHandler.FileService = server.FileService
	edgeStacksHandler.GitService = server.GitService

	var edgeTemplatesHandler = edgetemplates.NewHandler(requestBouncer)
	edgeTemplatesHandler.DataStore = server.DataStore

	var endpointHandler = endpoints.NewHandler(requestBouncer)
	endpointHandler.DataStore = server.DataStore
	endpointHandler.FileService = server.FileService
	endpointHandler.ProxyManager = server.ProxyManager
	endpointHandler.SnapshotService = server.SnapshotService
	endpointHandler.K8sClientFactory = server.KubernetesClientFactory
	endpointHandler.ReverseTunnelService = server.ReverseTunnelService
	endpointHandler.ComposeStackManager = server.ComposeStackManager
	endpointHandler.AuthorizationService = server.AuthorizationService

	var endpointEdgeHandler = endpointedge.NewHandler(requestBouncer)
	endpointEdgeHandler.DataStore = server.DataStore
	endpointEdgeHandler.FileService = server.FileService
	endpointEdgeHandler.ReverseTunnelService = server.ReverseTunnelService

	var endpointGroupHandler = endpointgroups.NewHandler(requestBouncer)
	endpointGroupHandler.AuthorizationService = server.AuthorizationService
	endpointGroupHandler.DataStore = server.DataStore

	var endpointProxyHandler = endpointproxy.NewHandler(requestBouncer)
	endpointProxyHandler.DataStore = server.DataStore
	endpointProxyHandler.ProxyManager = server.ProxyManager
	endpointProxyHandler.ReverseTunnelService = server.ReverseTunnelService

	var fileHandler = file.NewHandler(filepath.Join(server.AssetsPath, "public"))

	var motdHandler = motd.NewHandler(requestBouncer)

	var registryHandler = registries.NewHandler(requestBouncer)
	registryHandler.DataStore = server.DataStore
	registryHandler.FileService = server.FileService
	registryHandler.ProxyManager = server.ProxyManager
	registryHandler.K8sClientFactory = server.KubernetesClientFactory

	var resourceControlHandler = resourcecontrols.NewHandler(requestBouncer)
	resourceControlHandler.DataStore = server.DataStore

	var settingsHandler = settings.NewHandler(requestBouncer)
	settingsHandler.DataStore = server.DataStore
	settingsHandler.FileService = server.FileService
	settingsHandler.JWTService = server.JWTService
	settingsHandler.LDAPService = server.LDAPService
	settingsHandler.SnapshotService = server.SnapshotService

	var sslHandler = sslhandler.NewHandler(requestBouncer)
	sslHandler.SSLService = server.SSLService

	var stackHandler = stacks.NewHandler(requestBouncer)
	stackHandler.DataStore = server.DataStore
	stackHandler.DockerClientFactory = server.DockerClientFactory
	stackHandler.FileService = server.FileService
	stackHandler.SwarmStackManager = server.SwarmStackManager
	stackHandler.ComposeStackManager = server.ComposeStackManager
	stackHandler.KubernetesDeployer = server.KubernetesDeployer
	stackHandler.GitService = server.GitService

	var tagHandler = tags.NewHandler(requestBouncer)
	tagHandler.DataStore = server.DataStore

	var teamHandler = teams.NewHandler(requestBouncer)
	teamHandler.DataStore = server.DataStore

	var teamMembershipHandler = teammemberships.NewHandler(requestBouncer)
	teamMembershipHandler.DataStore = server.DataStore

	var statusHandler = status.NewHandler(requestBouncer, server.Status)

	var templatesHandler = templates.NewHandler(requestBouncer)
	templatesHandler.DataStore = server.DataStore
	templatesHandler.FileService = server.FileService
	templatesHandler.GitService = server.GitService

	var uploadHandler = upload.NewHandler(requestBouncer)
	uploadHandler.FileService = server.FileService

	var userHandler = users.NewHandler(requestBouncer, rateLimiter)
	userHandler.DataStore = server.DataStore
	userHandler.CryptoService = server.CryptoService

	var websocketHandler = websocket.NewHandler(requestBouncer)
	websocketHandler.DataStore = server.DataStore
	websocketHandler.SignatureService = server.SignatureService
	websocketHandler.ReverseTunnelService = server.ReverseTunnelService
	websocketHandler.KubernetesClientFactory = server.KubernetesClientFactory

	var webhookHandler = webhooks.NewHandler(requestBouncer)
	webhookHandler.DataStore = server.DataStore
	webhookHandler.DockerClientFactory = server.DockerClientFactory

	server.Handler = &handler.Handler{
		RoleHandler:            roleHandler,
		AuthHandler:            authHandler,
		BackupHandler:          backupHandler,
		CustomTemplatesHandler: customTemplatesHandler,
		EdgeGroupsHandler:      edgeGroupsHandler,
		EdgeJobsHandler:        edgeJobsHandler,
		EdgeStacksHandler:      edgeStacksHandler,
		EdgeTemplatesHandler:   edgeTemplatesHandler,
		EndpointGroupHandler:   endpointGroupHandler,
		EndpointHandler:        endpointHandler,
		EndpointEdgeHandler:    endpointEdgeHandler,
		EndpointProxyHandler:   endpointProxyHandler,
		FileHandler:            fileHandler,
		MOTDHandler:            motdHandler,
		RegistryHandler:        registryHandler,
		ResourceControlHandler: resourceControlHandler,
		SettingsHandler:        settingsHandler,
		SSLHandler:             sslHandler,
		StatusHandler:          statusHandler,
		StackHandler:           stackHandler,
		TagHandler:             tagHandler,
		TeamHandler:            teamHandler,
		TeamMembershipHandler:  teamMembershipHandler,
		TemplatesHandler:       templatesHandler,
		UploadHandler:          uploadHandler,
		UserHandler:            userHandler,
		WebSocketHandler:       websocketHandler,
		WebhookHandler:         webhookHandler,
	}

	handler := offlineGate.WaitingMiddleware(time.Minute, server.Handler)

	if server.HTTPEnabled {
		go func() {
			log.Printf("Starting HTTP server on port %s\n", server.BindAddress)
			httpServer := &http.Server{
				Addr:    server.BindAddress,
				Handler: handler,
			}

			go shutdown(server.ShutdownCtx, httpServer)
			err := httpServer.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Printf("[ERROR] [message: http server failed] [error: %s]", err)
			}
		}()
	}

	log.Printf("Starting HTTPS server on port %s\n", server.BindAddressHTTPS)
	httpsServer := &http.Server{
		Addr:    server.BindAddressHTTPS,
		Handler: handler,
	}

	httpsServer.TLSConfig = crypto.CreateServerTLSConfiguration()
	httpsServer.TLSConfig.GetCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		return server.SSLService.GetRawCertificate(), nil
	}

	go shutdown(server.ShutdownCtx, httpsServer)
	return httpsServer.ListenAndServeTLS("", "")
}

func shutdown(shutdownCtx context.Context, httpServer *http.Server) {
	<-shutdownCtx.Done()

	log.Println("[DEBUG] Shutting down http server")
	shutdownTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := httpServer.Shutdown(shutdownTimeout)
	if err != nil {
		fmt.Printf("Failed shutdown http server: %s \n", err)
	}
}
