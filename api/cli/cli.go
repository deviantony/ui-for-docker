package cli

import (
	"errors"
	"log"
	"time"

	"github.com/portainer/portainer/api"

	"os"
	"path/filepath"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Service implements the CLIService interface
type Service struct{}

var (
	errInvalidEndpointProtocol       = errors.New("Invalid endpoint protocol: Portainer only supports unix://, npipe:// or tcp://")
	errSocketOrNamedPipeNotFound     = errors.New("Unable to locate Unix socket or named pipe")
	errInvalidSnapshotInterval       = errors.New("Invalid snapshot interval")
	errAdminPassExcludeAdminPassFile = errors.New("Cannot use --admin-password with --admin-password-file")
)

// ParseFlags parse the CLI flags and return a portainer.Flags struct
func (*Service) ParseFlags(version string) (*portainer.CLIFlags, error) {
	kingpin.Version(version)

	flags := &portainer.CLIFlags{
		Addr:                      kingpin.Flag("bind", "Address and port to serve Portainer").Default(defaultBindAddress).Short('p').String(),
		TunnelAddr:                kingpin.Flag("tunnel-addr", "Address to serve the tunnel server").Default(defaultTunnelServerAddress).String(),
		TunnelPort:                kingpin.Flag("tunnel-port", "Port to serve the tunnel server").Default(defaultTunnelServerPort).String(),
		Assets:                    kingpin.Flag("assets", "Path to the assets").Default(defaultAssetsDirectory).Short('a').String(),
		Data:                      kingpin.Flag("data", "Path to the folder where the data is stored").Default(defaultDataDirectory).Short('d').String(),
		EndpointURL:               kingpin.Flag("host", "Endpoint URL").Short('H').String(),
		EnableEdgeComputeFeatures: kingpin.Flag("edge-compute", "Enable Edge Compute features").Bool(),
		NoAnalytics:               kingpin.Flag("no-analytics", "Disable Analytics in app (deprecated)").Bool(),
		TLS:                       kingpin.Flag("tlsverify", "TLS support").Default(defaultTLS).Bool(),
		TLSSkipVerify:             kingpin.Flag("tlsskipverify", "Disable TLS server verification").Default(defaultTLSSkipVerify).Bool(),
		TLSCacert:                 kingpin.Flag("tlscacert", "Path to the CA").Default(defaultTLSCACertPath).String(),
		TLSCert:                   kingpin.Flag("tlscert", "Path to the TLS certificate file").Default(defaultTLSCertPath).String(),
		TLSKey:                    kingpin.Flag("tlskey", "Path to the TLS key").Default(defaultTLSKeyPath).String(),
		SSL:                       kingpin.Flag("ssl", "Secure Portainer instance using SSL").Default(defaultSSL).Bool(),
		SSLCert:                   kingpin.Flag("sslcert", "Path to the SSL certificate used to secure the Portainer instance").Default(defaultSSLCertPath).String(),
		SSLKey:                    kingpin.Flag("sslkey", "Path to the SSL key used to secure the Portainer instance").Default(defaultSSLKeyPath).String(),
		SnapshotInterval:          kingpin.Flag("snapshot-interval", "Duration between each endpoint snapshot job").Default(defaultSnapshotInterval).String(),
		AdminPassword:             kingpin.Flag("admin-password", "Hashed admin password").String(),
		AdminPasswordFile:         kingpin.Flag("admin-password-file", "Path to the file containing the password for the admin user").String(),
		Labels:                    pairs(kingpin.Flag("hide-label", "Hide containers with a specific label in the UI").Short('l')),
		Logo:                      kingpin.Flag("logo", "URL for the logo displayed in the UI").String(),
		Templates:                 kingpin.Flag("templates", "URL to the templates definitions.").Short('t').String(),
		JsonLogger:                kingpin.Flag("json-logger", "Use JSON logger if you do log aggregation with tools like Splunk or Logstash.").Bool(),
	}

	kingpin.Parse()

	if !filepath.IsAbs(*flags.Assets) {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		*flags.Assets = filepath.Join(filepath.Dir(ex), *flags.Assets)
	}

	return flags, nil
}

// ValidateFlags validates the values of the flags.
func (*Service) ValidateFlags(flags *portainer.CLIFlags) error {

	displayDeprecationWarnings(flags)

	err := validateEndpointURL(*flags.EndpointURL)
	if err != nil {
		return err
	}

	err = validateSnapshotInterval(*flags.SnapshotInterval)
	if err != nil {
		return err
	}

	if *flags.AdminPassword != "" && *flags.AdminPasswordFile != "" {
		return errAdminPassExcludeAdminPassFile
	}

	return nil
}

func displayDeprecationWarnings(flags *portainer.CLIFlags) {
	if *flags.NoAnalytics {
		log.Println("Warning: The --no-analytics flag has been kept to allow migration of instances running a previous version of Portainer with this flag enabled, to version 2.0 where enabling this flag will have no effect.")
	}
}

func validateEndpointURL(endpointURL string) error {
	if endpointURL != "" {
		if !strings.HasPrefix(endpointURL, "unix://") && !strings.HasPrefix(endpointURL, "tcp://") && !strings.HasPrefix(endpointURL, "npipe://") {
			return errInvalidEndpointProtocol
		}

		if strings.HasPrefix(endpointURL, "unix://") || strings.HasPrefix(endpointURL, "npipe://") {
			socketPath := strings.TrimPrefix(endpointURL, "unix://")
			socketPath = strings.TrimPrefix(socketPath, "npipe://")
			if _, err := os.Stat(socketPath); err != nil {
				if os.IsNotExist(err) {
					return errSocketOrNamedPipeNotFound
				}
				return err
			}
		}
	}
	return nil
}

func validateSnapshotInterval(snapshotInterval string) error {
	if snapshotInterval != defaultSnapshotInterval {
		_, err := time.ParseDuration(snapshotInterval)
		if err != nil {
			return errInvalidSnapshotInterval
		}
	}
	return nil
}
