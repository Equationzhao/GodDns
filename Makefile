App = GodDns

GOPATH=$(shell go env GOPATH)
OS=$(shell go env GOOS)
Linux = linux
Windows = windows

all: tool check clean build

fmt: ## Format the code
	$(info Formatting the code)
	go fmt ./...

vet: ## Vet the code
	$(info Vet the code)
	go vet ./...

lint:
	golangci-lint run

check: fmt vet lint ## Run all the checks
	go mod tidy

tool: ## Install the tools
	$(info Installing the tools)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest


build: ## Build the binary
	$(info Building the binary)
	@mkdir "build"
	@if [ ${OS} = ${Linux} ]; \
	then \
		go build -ldflags="-s -w" -o build/${App} GodDns/Cmd/GodDns; \
	elif [ ${OS} = ${Windows} ]; \
    then \
		go build -ldflags="-s -w" -o build/${App}.exe GodDns/Cmd/GodDns; \
  	else \
  		echo "Unsupported OS"; \
  		echo "Please remove the binary manually"; \
  		echo "Path: $(GOPATH)/bin/${App}"; \
  	fi \


init: ## Initialize the config
	$(info Initializing the config)
	go run GodDns/Cmd/GodDns g

run race: ## Run the binary, checking date race
	$(info Running the binary)
	go run -race GodDns/Cmd/GodDns run auto -parallel

clean: ## Clean up the build
	rm -rf build
	go clean

install: ## Install the binary to the GOPATH
	$(info Installing the binary)
	go install GodDns/Cmd/GodDns

uninstall : ## Uninstall the binary from GOPATH
	@if [ ${OS} = ${Linux} ]; \
	then \
	  	echo "Uninstalling the binary" "from" "$(GOPATH)/bin";\
		rm -f $(GOPATH)/bin/${App}; \
	elif [ ${OS} = ${Windows} ]; \
    then \
    	echo "Uninstalling the binary" "from" "${GOPATH}\\bin"; \
		rm -f "${GOPATH}\\bin\\${App}.exe"; \
  	else \
  		echo "Unsupported OS"; \
  		echo "Please remove the binary manually"; \
  		echo "Path: $(GOPATH)/bin/${App}"; \
  	fi \
