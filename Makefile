container_runtime := $(shell which podman || which docker)

$(info using ${container_runtime})

build-images:
	cd search-services && ${container_runtime} build -f Dockerfile.api -t api:latest .
	cd search-services && ${container_runtime} build -f Dockerfile.frontend -t frontend:latest .
	cd search-services && ${container_runtime} build -f Dockerfile.words -t words:latest .
	cd search-services && ${container_runtime} build -f Dockerfile.update -t update:latest .
	cd search-services && ${container_runtime} build -f Dockerfile.search -t search:latest .
	cd search-services && ${container_runtime} build -f Dockerfile.bot -t bot:latest .
	cd tests && ${container_runtime} build -t tests:latest .

up: down build-images
	${container_runtime} compose up -d

down:
	${container_runtime} compose down

clean:
	${container_runtime} compose down -v

run-tests: 
	${container_runtime} run --rm --network=host tests:latest

test:
	make clean
	make up
	@echo wait cluster to start && sleep 10
	make run-tests
	make clean
	@echo "test finished"

lint:
	make -C search-services lint

unit:
	make -C search-services unit

proto:
	make -C search-services protobuf

tools:
	go install github.com/yoheimuta/protolint/cmd/protolint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.4.0
	@echo "checking protobuf compiler, if it fails follow guide at https://protobuf.dev/installation/"
	@which -s protoc && echo OK || exit 1
