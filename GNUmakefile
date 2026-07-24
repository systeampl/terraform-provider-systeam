default: build

build:
	go build -o terraform-provider-systeam

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/pawel-cygal/systeam/0.1.0/linux_amd64
	cp terraform-provider-systeam ~/.terraform.d/plugins/registry.terraform.io/pawel-cygal/systeam/0.1.0/linux_amd64/

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v -timeout 120m

lint:
	golangci-lint run ./...

.PHONY: default build install test testacc lint
