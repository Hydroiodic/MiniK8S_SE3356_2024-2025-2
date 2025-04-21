# TODO
# Install containerd
install-containerd:
	@sudo $(MAKE) -f scripts/containerd.mk

install-golangci:
	$(MAKE) -f scripts/golangci-lint.mk

install-golines:
	$(MAKE) -f scripts/golines.mk install

clean:
	@rm -rf build/*
