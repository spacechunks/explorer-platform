WORKDIR := work
CNI_PLUGINS := $(WORKDIR)/plugins/bin
SUDO := $(sudo --preserve-env=PATH,CNI_PATH env)


.PHONY: setup
setup:
	$(SUDO) apt update
	$(SUDO) apt install -y linux-tools-common libbpf-dev
	$(SUDO) mount bpffs /sys/fs/bpf -t bpf

.PHONY: vmlinux
vmlinux:
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > internal/tun/bpf/include/vmlinux.h

.PHONY: gogen_all
gogen_all:
	go generate ./...

.PHONY: genproto
genproto:
	buf generate --template ./api/buf.gen.yaml --output ./api ./api

.PHONY: nodedev
nodedev:
	./nodedev/up.sh

.PHONY: e2etests
e2etests:
	GOOS=linux GOARCH=arm64 go build -o ./nodedev/ptpnat ./cmd/ptpnat/main.go
	$(SUDO) go test ./test/e2e/...

# functests require CNI_PATH to be set
export CNI_PATH=$(shell pwd)/$(CNI_PLUGINS)

functests: $(CNI_PLUGINS)
	$(SUDO) go test ./test/functional/...

$(CNI_PLUGINS): $(WORKDIR)
	git clone git@github.com:containernetworking/plugins.git $(CNI_PLUGINS)
	$(CNI_PLUGINS)/build_linux.sh

$(WORKDIR):
	mkdir $(WORKDIR)
