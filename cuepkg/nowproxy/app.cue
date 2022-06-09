package nowproxy

import (
	"github.com/innoai-tech/runtime/cuepkg/kube"
)

#NowProxy: kube.#App & {
	app: {
		name:    _ | *"now-proxy"
		version: _ | *"debug2"
	}

	services: "\(app.name)": {
		selector: "app": app.name
		ports:  containers."now-proxy".ports
		expose: _ | *{
			host: string | *"now-proxy.x.io"
			paths: http: "/"
		}
	}

	containers: "now-proxy": {
		image: {
			name: _ | *"ghcr.io/morlay/now-proxy"
			tag:  _ | *"\(app.version)"
		}
		ports: http: 80
		env: PORT:   "80"
		readinessProbe: kube.#ProbeHttpGet & {
			httpGet: {path: "/http:/www.gstatic.com/generate_204", port: ports.http}
		}
		livenessProbe: readinessProbe
	}
}
