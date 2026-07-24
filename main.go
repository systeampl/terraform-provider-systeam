package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/pawel-cygal/terraform-provider-systeam/internal/provider"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/pawel-cygal/systeam",
	})
	if err != nil {
		log.Fatal(err)
	}
}
