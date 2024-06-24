WORKDIR := work
CNI_PLUGINS := $(WORKDIR)/plugins

.PHONY: setup
setup:
	apt update
	apt install -y linux-tools-common libbpf-dev

.PHONY: vmlinux
vmlinux:
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > internal/tun/bpf/include/vmlinux.h

.PHONY: gogen_all
gogen_all:
	go generate ./...

export CNI_PATH=$(CNI_PLUGINS)

.PHONY: functests
functests: $(CNI_PLUGINS)
	go test ./test/functional/...

$(CNI_PLUGINS): $(WORKDIR)
	git clone git@github.com:containernetworking/plugins.git $(WORKDIR)
	$(WORKDIR)/build_linux.sh

$(WORKDIR):
	mkdir $(WORKDIR)
