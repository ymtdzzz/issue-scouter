.PHONY: help
help: ## Print help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-lint-tools: ## Install lint tools
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/alexkohler/prealloc@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

DIR=./...
lint: ## Run static analysis
	go vet "$(DIR)"
	test -z "`gofmt -s -d .`"
	staticcheck "$(DIR)"
	prealloc -set_exit_status "$(DIR)"
	gosec -conf gosec.json "$(DIR)"

.PHONY: test
test: ## Run test ex.) make test OPT="-run TestXXX"
	go test ./pkg/... ./cmd/... -v "$(OPT)"

test-coverage: ## Run test with coverage
	$(MAKE) test OPT="-coverprofile=coverage.out"
	go tool cover -html=coverage.out

run-local-action: ## Run action command locally. gh command is needed
	@GITHUB_TOKEN=$(shell gh auth token) INPUT_CONFIG_FILE="./example.yml" go run ./cmd/action

run-local-labels: ## Run labels command locally. gh command is needed
	@GITHUB_TOKEN=$(shell gh auth token) INPUT_CONFIG_FILE="./example.yml" go run ./cmd/labels
