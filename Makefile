
install: install-golangci install-golines

install-golangci:
	$(MAKE) -f scripts/golangci-lint.mk

install-golines:
	$(MAKE) -f scripts/golines.mk install

clean:
	@rm -rf build/*

lint:
	@echo "Running golangci-lint..."
	@golangci-lint-v2 run

test:
	@echo "Running tests..."
	@go test -v ./test/...
	@echo "Tests completed."
	@echo "Check test.log for details."
	@echo "Check coverage.out for coverage details."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: install-golangci install-golines clean lint test
