package main

//go:generate dagger mod sync

import (
	"context"
	"fmt"
)

type SampleApp struct{}

const (
	DefaultRegistry         string = "ghcr.io/SaimonWoidig"
	DefaultBackendImageName string = "backend"
)

func GetAuthdContainer(registryUser string, registryPassword string, registryPath Optional[string]) *Container {
	return dag.Container().WithRegistryAuth(fmt.Sprintf("%v/%v", DefaultRegistry, DefaultBackendImageName), registryUser, dag.SetSecret("registryPassword", registryPassword))
}

func (m *SampleApp) BuildBackend(ctx context.Context) *Container {
	backendSource := dag.Host().Directory("./backend")
	authdCtr := dag.Container().
		WithRegistryAuth("docker.io", "", dag.SetSecret("", ""))
	return dag.
		Golang().
		WithContainer(authdCtr).
		WithVersion("1.21.4").
		WithSource(backendSource).
		Download().
		Build()
}
