App = GodDns

all: tool check clean build

fmt: ## Format the code
	$(info Formatting the code)
	go fmt ./...

vet: ## Vet the code
	$(info Vet the code)
	go vet ./...

lint: ## Lint the code
	$(info Lint the code)
	golangci-lint run

check: fmt vet lint ## Run all the checks

tool: ## Install the tools
	$(info Installing the tools)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

build: ## Build the binary
	$(info Building the binary)
	mkdir "build"
	go build -ldflags="-s -w" -o build/${App} GodDns/Cmd
	go build -ldflags="-s -w" -o build/${App}.exe GodDns/Cmd

run: ## Run the binary
	$(info Running the binary)
	go run -race GodDns/Cmd run auto -parallel

clean: ## Clean up the build
	rm -rf build
