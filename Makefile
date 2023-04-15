App = GodDns

GOPATH=$(shell go env GOPATH)
OS=$(shell go env GOOS)
Linux = linux
Windows = windows

.PHONY: pre
pre: tool

.PHONY: all
all: check clean build

.PHONY: fmt
fmt: ## Format the code
	$(info Formatting the code)
	gofumpt -l -w .

.PHONY : vet
vet: ## Vet the code
	$(info Vet the code)
	go vet ./...

.PHONY : lint
lint:
	golangci-lint run

.PHONY: mod
mod:
	go mod tidy

.PHONY: check
check: mod fmt vet lint gci ## Run all the checks

.PHONY: gci
gci: ## Run gci
	$(info Running gci)
	gci write .

.PHONY: tool
tool: ## Install the tools
	$(info Installing the tools)
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest


build: ## Build the binary
	$(info Building the binary)
	@mkdir "build"
	@if [ ${OS} = ${Linux} ]; \
	then \
		go build -ldflags="-s -w" -o build/${App} GodDns/cmd/GodDns; \
	elif [ ${OS} = ${Windows} ]; \
    then \
		go build -ldflags="-s -w" -o build/${App}.exe GodDns/cmd/GodDns; \
  	else \
  		echo "Unsupported OS"; \
  		echo "Please remove the binary manually"; \
  		echo "Path: $(GOPATH)/bin/${App}"; \
  	fi \

rebuild: clean build ## Clean and build the binary

.PHONY: init
init: ## Initialize the config
	$(info Initializing the config)
	go run GodDns/cmd/GodDns g

.PHONY: run
run race: ## Run the binary, checking date race
	$(info Running the binary)
	go run -race GodDns/cmd/GodDns run auto -parallel

.PHONY: clean
clean: ## Clean up the build
	rm -rf build
	go clean

.PHONY : install
install: ## Install the binary to the GOPATH
	$(info Installing the binary)
	go install GodDns/cmd/GodDns

.PHONY : uninstall
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


upx: ## Compress the binary
	$(info Compressing the binary)
	@if [ ${OS} = ${Linux} ]; \
	then \
	  	upx build/${App}; \
	elif [ ${OS} = ${Windows} ]; \
    then \
	  	upx build/${App}.exe; \
  	fi \


build-all: ## Build the binary for all the platforms
	$(info Building the binary for all the platforms)
	@mkdir "build"
	@echo "Building for Windows"
	@GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/${App}-Windows-amd64.exe GodDns/cmd/GodDns
	@echo "Building for Linux"
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/${App}-Linux-amd64 GodDns/cmd/GodDns
