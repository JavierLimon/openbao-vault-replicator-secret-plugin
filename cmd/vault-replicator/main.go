package main

import (
	"fmt"
	"os"

	"github.com/openbao/openbao/api/v2"
	"github.com/openbao/openbao/sdk/v2/plugin"

	replicator "github.com/JavierLimon/openbao-vault-replicator-secret-plugin/plugin"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("OpenBAO Vault Replicator\n")
		fmt.Printf("Version: %s\n", replicator.GetVersion())
		fmt.Printf("Commit: %s\n", replicator.GetCommit())
		fmt.Printf("Date: %s\n", replicator.GetDate())
		fmt.Printf("BuildType: %s\n", replicator.GetBuildType())
		os.Exit(0)
	}

	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse flags: %s\n", err)
		os.Exit(1)
	}

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	err := plugin.ServeMultiplex(&plugin.ServeOpts{
		BackendFactoryFunc: replicator.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	})
	if err != nil {
		os.Exit(1)
	}
}
