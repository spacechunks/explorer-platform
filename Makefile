WORKDIR := work
CNI_PLUGINS := $(WORKDIR)/plugins/bin

.PHONY: setup
setup:
	sudo apt update
	sudo apt install -y linux-tools-common libbpf-dev
	sudo mount bpffs /sys/fs/bpf -t bpf

.PHONY: vmlinux
vmlinux:
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > internal/tun/bpf/include/vmlinux.h

.PHONY: gogen_all
gogen_all:
	go generate ./...

export CNI_PATH=$(shell pwd)/$(CNI_PLUGINS)

.PHONY: nodedev
nodedev:
	./nodedev/up.sh

functests: $(CNI_PLUGINS)
	sudo --preserve-env=PATH,CNI_PATH env go test ./test/functional/...

$(CNI_PLUGINS): $(WORKDIR)
	git clone git@github.com:containernetworking/plugins.git $(CNI_PLUGINS)
	$(CNI_PLUGINS)/build_linux.sh

$(WORKDIR):
	mkdir $(WORKDIR)
