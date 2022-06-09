package main

import (
	"dagger.io/dagger"
	"dagger.io/dagger/core"
	"universe.dagger.io/docker"

	"github.com/innoai-tech/runtime/cuepkg/debian"
	"github.com/innoai-tech/runtime/cuepkg/golang"
	"github.com/innoai-tech/runtime/cuepkg/tool"
)

dagger.#Plan & {
	client: {
		env: {
			VERSION: string | *"dev"

			GIT_SHA: string | *""
			GIT_REF: string | *""

			GOPROXY:   string | *""
			GOPRIVATE: string | *""
			GOSUMDB:   string | *""

			GH_USERNAME: string | *""
			GH_PASSWORD: dagger.#Secret

			LINUX_MIRROR: string | *""
		}
	}

	actions: {
		_source: core.#Source & {
			path: "."
			include: [
				"cmd/",
				"pkg/",
				"go.mod",
				"go.sum",
			]
		}

		_version: (tool.#ResolveVersion & {ref: client.env.GIT_REF, version: "\(client.env.VERSION)"}).output

		_tag: _version

		_archs: ["amd64", "arm64"]

		info: golang.#Info & {
			source: _source.output
		}

		build: golang.#Build & {
			source: _source.output
			image: mirror: client.env.LINUX_MIRROR
			go: {
				os: ["linux"]
				arch:    _archs
				package: "./cmd/now-proxy"
				ldflags: [
					"-s -w",
					"-X \(info.module)/pkg/version.Version=\(_version)",
					"-X \(info.module)/pkg/version.Revision=\(client.env.GIT_SHA)",
				]
			}
			run: env: {
				GOPROXY:   client.env.GOPROXY
				GOPRIVATE: client.env.GOPRIVATE
				GOSUMDB:   client.env.GOSUMDB
			}
		}

		images: {
			for _arch in _archs {
				"linux/\(_arch)": docker.#Build & {
					steps: [
						debian.#Build & {
							mirror: client.env.LINUX_MIRROR
							packages: {
								"ca-certificates": _
							}
						},
						docker.#Copy & {
							contents: build["linux/\(_arch)"].output
							dest:     "/\(build.go.name)"
							source:   "./\(build.go.name)"
						},
						docker.#Set & {
							config: {
								label: {
									"org.opencontainers.image.source":   "https://\(info.module)"
									"org.opencontainers.image.revision": "\(client.env.GIT_SHA)"
								}
								workdir: "/"
								env: PORT: "80"
								entrypoint: ["/\(build.go.name)"]
							}
						},
					]
				}
			}
		}

		push: docker.#Push & {
			"dest": "ghcr.io/morlay/now-proxy:\(_tag)"
			"images": {
				for p, i in images {
					"\(p)": i.output
				}
			}
			"auth": {
				username: client.env.GH_USERNAME
				secret:   client.env.GH_PASSWORD
			}
		}
	}
}
