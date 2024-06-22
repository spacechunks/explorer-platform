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
