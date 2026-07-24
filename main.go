package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/systeampl/terraform-provider-systeam/internal/provider"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/systeampl/systeam",
	})
	if err != nil {
		log.Fatal(err)
	}
}
