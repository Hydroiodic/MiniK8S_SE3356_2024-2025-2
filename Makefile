# TODO
# Install containerd
install-containerd:
	sudo make -f scripts/containerd.mak

clean:
	rm -rf build/*