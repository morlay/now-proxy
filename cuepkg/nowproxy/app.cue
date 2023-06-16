package nowproxy

import (
	kubepkg "github.com/octohelm/kubepkg/cuepkg/kubepkg"
)

#NowProxy: kubepkg.#KubePkg & {
	metadata: {
		name: _ | *"now-proxy"
	}

	spec: {
		version: _ | *"debug2"

		deploy: {
			kind: "Deployment"
			spec: {
				replicas: _ | *1
			}
		}

		services: "#": {
			ports: containers."now-proxy".ports
			paths: http: "/"
			expose: _ | *{
				type:    "Ingress"
				gateway: _ | *["now-proxy.x.io"]
			}
		}

		containers: "now-proxy": {
			image: {
				name: _ | *"ghcr.io/morlay/now-proxy"
				tag:  _ | *"\(spec.version)"
			}
			ports: http: 80
			env: PORT:   "80"
			readinessProbe: kubepkg.#Probe & {
				httpGet: {path: "/http:/www.gstatic.com/generate_204", port: ports.http}
			}
			livenessProbe: readinessProbe
		}
	}
}
