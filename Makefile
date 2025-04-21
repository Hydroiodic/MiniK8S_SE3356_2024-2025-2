# TODO
# Install containerd
install-containerd:
	@sudo $(MAKE) -f scripts/containerd.mk

install-golangci:
	$(MAKE) -f scripts/golangci-lint.mk

clean:
	@rm -rf build/*
