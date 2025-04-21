# Makefile for installing and configuring containerd

# Default versions
CONTAINERD_VERSION ?= 1.6.2
RUNC_VERSION ?= 1.1.12
CNI_VERSION ?= 1.1.1

# Architecture (default: amd64)
ARCH ?= amd64

# Installation paths
CONTAINERD_URL = https://github.com/containerd/containerd/releases/download/v$(CONTAINERD_VERSION)/containerd-$(CONTAINERD_VERSION)-linux-$(ARCH).tar.gz
RUNC_URL = https://github.com/opencontainers/runc/releases/download/v$(RUNC_VERSION)/runc.$(ARCH)
CNI_URL = https://github.com/containernetworking/plugins/releases/download/v$(CNI_VERSION)/cni-plugins-linux-$(ARCH)-v$(CNI_VERSION).tgz
SYSTEMD_SERVICE_URL = https://raw.githubusercontent.com/containerd/containerd/main/containerd.service

# Temporary directory for downloads
TMP_DIR := /tmp/containerd-install

# Default target
.PHONY: all
all: install-containerd install-runc install-cni configure-systemd

# Create temporary directory
.PHONY: setup
setup:
	@mkdir -p $(TMP_DIR)
	@mkdir -p /opt/cni/bin
	@mkdir -p /usr/local/lib/systemd/system

# Install containerd
.PHONY: install-containerd
install-containerd: setup
	@echo "Downloading containerd $(CONTAINERD_VERSION)..."
	@curl -sL $(CONTAINERD_URL) -o $(TMP_DIR)/containerd.tar.gz
	# @echo "Verifying containerd checksum..."
	# @curl -sL $(CONTAINERD_URL).sha256sum | sha256sum -c || { echo "Checksum verification failed"; exit 1; }
	@echo "Extracting containerd..."
	@tar Cxzvf /usr/local $(TMP_DIR)/containerd.tar.gz
	@echo "containerd installed successfully."

# Install runc
.PHONY: install-runc
install-runc: setup
	@echo "Downloading runc $(RUNC_VERSION)..."
	@curl -sL $(RUNC_URL) -o $(TMP_DIR)/runc
	# @echo "Verifying runc checksum..."
	# @curl -sL $(RUNC_URL).sha256sum | sha256sum -c || { echo "Checksum verification failed"; exit 1; }
	@echo "Installing runc..."
	@install -m 755 $(TMP_DIR)/runc /usr/local/sbin/runc
	@echo "runc installed successfully."

# Install CNI plugins
.PHONY: install-cni
install-cni: setup
	@echo "Downloading CNI plugins $(CNI_VERSION)..."
	@curl -sL $(CNI_URL) -o $(TMP_DIR)/cni-plugins.tgz
	# @echo "Verifying CNI plugins checksum..."
	# @curl -sL $(CNI_URL).sha256sum | sha256sum -c || { echo "Checksum verification failed"; exit 1; }
	@echo "Extracting CNI plugins..."
	@tar Cxzvf /opt/cni/bin $(TMP_DIR)/cni-plugins.tgz
	@echo "CNI plugins installed successfully."

# Configure systemd
.PHONY: configure-systemd
configure-systemd: setup
	@echo "Downloading containerd systemd service file..."
	@curl -sL $(SYSTEMD_SERVICE_URL) -o /usr/local/lib/systemd/system/containerd.service
	@echo "Reloading systemd daemon..."
	@systemctl daemon-reload
	@echo "Enabling and starting containerd service..."
	@systemctl enable --now containerd
	@echo "containerd service configured and started."

# Generate default configuration
.PHONY: configure
configure:
	@echo "Generating default containerd configuration..."
	@mkdir -p /etc/containerd
	@containerd config default > /etc/containerd/config.toml
	@echo "Default configuration generated at /etc/containerd/config.toml."
	@echo "Please review and adjust the configuration as needed."

# Clean up temporary files
.PHONY: clean
clean:
	@echo "Cleaning up temporary files..."
	@rm -rf $(TMP_DIR)
	@echo "Cleanup complete."

# Uninstall containerd, runc, and CNI plugins
.PHONY: uninstall
uninstall:
	@echo "Stopping containerd service..."
	@systemctl stop containerd || true
	@echo "Disabling containerd service..."
	@systemctl disable containerd || true
	@echo "Removing containerd binaries..."
	@rm -f /usr/local/bin/containerd*
	@echo "Removing runc binary..."
	@rm -f /usr/local/sbin/runc
	@echo "Removing CNI plugins..."
	@rm -rf /opt/cni/bin/*
	@echo "Removing systemd service file..."
	@rm -f /usr/local/lib/systemd/system/containerd.service
	@echo "Reloading systemd daemon..."
	@systemctl daemon-reload
	@echo "Removing containerd configuration..."
	@rm -rf /etc/containerd
	@echo "Uninstallation complete."