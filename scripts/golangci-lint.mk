
install:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.2
	@echo "The version of golangci-lint is v2, so please change your VSCode plugins \"golang.go\" to pre-release version."
