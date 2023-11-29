package main

//go:generate dagger mod sync

import (
	"context"
	"fmt"
	"strconv"
)

type Ci struct{}

const (
	ImageSource             string = "https://github.com/SaimonWoidig/dagger-ci"
	DefaultRegistry         string = "ghcr.io/saimonwoidig/dagger-ci"
	DefaultBackendImageName string = "backend"
	DefaultGolangVersion    string = "1.21.4"
	DefaultBackendPort      int    = 8080
)

func (m *Ci) golangBuilder() *Container {
	return dag.Container().
		WithRegistryAuth("docker.io", "", dag.SetSecret("", "")).
		From(fmt.Sprintf("docker.io/library/golang:%v", DefaultGolangVersion)).
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
		WithExec([]string{"build", "-o", "/out/backend", "-ldflags", "-extldflags '-static'", "-tags", "osusergo,netgo"}).
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
