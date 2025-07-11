// main.go
package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	cloudflareplus "github.com/jgalluzzi/cloudflare_plus"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: cloudflareplus.Provider,
	})
}
