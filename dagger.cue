package main

import (
	"strings"
	"dagger.io/dagger"

	"github.com/innoai-tech/runtime/cuepkg/tool"
	"github.com/innoai-tech/runtime/cuepkg/golang"
)

dagger.#Plan

client: env: {
	VERSION: string | *"dev"
	GIT_SHA: string | *""
	GIT_REF: string | *""

	GOPROXY:   string | *""
	GOPRIVATE: string | *""
	GOSUMDB:   string | *""

	GH_USERNAME: string | *""
	GH_PASSWORD: dagger.#Secret

	CONTAINER_REGISTRY_PULL_PROXY: string | *""
	LINUX_MIRROR:                  string | *""
}

client: filesystem: {
	"build/output": write: contents: actions.go.archive.output
}

actions: version: tool.#ResolveVersion & {
	"ref":     "\(client.env.GIT_REF)"
	"version": "\(client.env.VERSION)"
}

mirror: {
	linux: "\(client.env.LINUX_MIRROR)"
	pull:  "\(client.env.CONTAINER_REGISTRY_PULL_PROXY)"
}

actions: go: golang.#Project & {
	source: {
		path: "."
		include: [
			"cmd/",
			"pkg/",
			"go.mod",
			"go.sum",
		]
	}

	version:  "\(actions.version.output)"
	revision: "\(client.env.GIT_SHA)"

	goos: ["linux"]
	goarch: ["amd64", "arm64"]
	main: "./cmd/now-proxy"
	ldflags: [
		"-s -w",
		"-X \(go.module)/pkg/version.Version=\(go.version)",
		"-X \(go.module)/pkg/version.Revision=\(go.revision)",
	]

	env: {
		GOPROXY:   client.env.GOPROXY
		GOPRIVATE: client.env.GOPRIVATE
		GOSUMDB:   client.env.GOSUMDB
	}

	build: {
		image: {
			"mirror": mirror
		}
	}

	ship: {
		name: "\(strings.Replace(actions.go.module, "github.com/", "ghcr.io/", -1))"
		tag:  "\(actions.version.output)"

		image: {
			source:   "gcr.io/distroless/static-debian11:debug"
			"mirror": mirror
		}

		config: env: {
			PORT: "80"
		}

		push: {
			auth: {
				username: client.env.GH_USERNAME
				secret:   client.env.GH_PASSWORD
			}
		}
	}
}
