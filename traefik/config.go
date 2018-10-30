package traefik

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"
)

type traefikFrontend map[string]*frontendConfig

type frontendConfig struct {
	Backend        string `json:"backend"`
	PassHostHeader bool   `json:"passHostHeader"`
}

type traefikBackend map[string]*backendConfig

type backendConfig map[string]interface{}

type traefikBackendServer map[string]*backendServerConfig

type backendServerConfig struct {
	URL string
}

type stickyConfig struct {
	CookieName string `json:"cookieName"`
}

type loadBalancerConfig struct {
	Stickiness *stickyConfig `json:"stickiness"`
	Method     string        `json:"method"`
}

type traefikConfig map[string]interface{}

func NewJsonConfig(frontendName string, backendPort string, backends []net.IP) ([]byte, error) {
	traefikBackendServers := make(traefikBackendServer)
	for idx, ip := range backends {
		backendServerName := strings.Join([]string{"backend", strconv.Itoa(idx)}, "")
		traefikBackendServers[backendServerName] = &backendServerConfig{
			URL: "http://" + ip.String() + ":" + backendPort,
		}
	}
	backendName := strings.Join([]string{frontendName, strconv.Itoa(1)}, "")
	configTest := &traefikConfig{
		"frontends": &traefikFrontend{frontendName: &frontendConfig{Backend: backendName, PassHostHeader: true}},
		"backends": &traefikBackend{backendName: &backendConfig{
			"servers": &traefikBackendServers,
			"loadbalancer": &loadBalancerConfig{
				Method: "drr",
				Stickiness: &stickyConfig{
					CookieName: "__traefik",
				},
			},
		},
		},
	}

	return json.Marshal(configTest)
}
