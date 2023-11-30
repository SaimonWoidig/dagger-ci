package main

//go:generate dagger mod sync

import (
	"context"
	"fmt"
	"strconv"
)

type Ci struct{}

const (
	ImageSource     string = "https://github.com/SaimonWoidig/dagger-ci"
	DefaultRegistry string = "ghcr.io/saimonwoidig/dagger-ci"

	DefaultBackendImageName string = "backend"
	DefaultGolangVersion    string = "1.21.4"
	DefaultBackendPort      int    = 8080

	DefaultNginxVersion string = "mainline-alpine-slim"
	DefaultFrontendPort int    = 8080
)

func (m *Ci) golangBuilder() *Container {
	golangImageSource := "docker.io/library/golang"
	return dag.Container().
		From(fmt.Sprintf("%v:%v", golangImageSource, DefaultGolangVersion)).
		WithWorkdir("/src").
		WithEntrypoint([]string{"go"}).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("gomod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("gobuild"))
}

func (m *Ci) BuildBackend() *File {
	src := dag.Host().Directory("./backend")
	return m.golangBuilder().
		WithMountedDirectory("/src", src).
		WithExec([]string{"mod", "download"}).
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"build", "-v", "-o", "/out/backend", "-ldflags", "-v -s -w -extldflags '-static'", "-tags", "osusergo,netgo", "-trimpath"}).
		Directory("/out").
		File("backend")
}

func (m *Ci) BackendImage() *Container {
	return dag.Container().
		WithWorkdir("/").
		WithFile("/backend", m.BuildBackend()).
		WithLabel("org.opencontainers.image.source", ImageSource).
		WithEnvVariable("PORT", strconv.Itoa(DefaultBackendPort)).
		WithExposedPort(DefaultBackendPort).
		WithEntrypoint([]string{"/backend"})
}

func (m *Ci) PublishBackend(ctx context.Context, regUser string, regPass string, imageTag Optional[string]) (string, error) {
	tag := imageTag.GetOr("develop")
	fullImageRef := fmt.Sprintf("%v/%v:%v", DefaultRegistry, DefaultBackendImageName, tag)
	return m.BackendImage().
		WithRegistryAuth(DefaultRegistry, regUser, dag.Host().SetSecretFile("registryPassword", regPass)).
		Publish(ctx, fullImageRef)
}

func (m *Ci) FrontendImage() *Container {
	nginxImageSource := "docker.io/nginxinc/nginx-unprivileged"
	src := dag.Host().Directory("./frontend")
	return dag.Container().
		From(fmt.Sprintf("%v:%v", nginxImageSource, DefaultNginxVersion)).
		WithLabel("org.opencontainers.image.source", ImageSource).
		WithDirectory("/usr/share/nginx/html", src).
		WithExposedPort(DefaultFrontendPort)
}
