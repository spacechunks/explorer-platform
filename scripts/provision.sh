#!/bin/bash
apt update
apt-get install -y gnupg2

# crio
MAJOR_VERSION=1.30
curl -fsSL https://pkgs.k8s.io/addons:/cri-o:/stable:/v$MAJOR_VERSION/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/cri-o-apt-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/cri-o-apt-keyring.gpg] https://pkgs.k8s.io/addons:/cri-o:/stable:/v$MAJOR_VERSION/deb/ /" | tee /etc/apt/sources.list.d/cri-o.list
apt-get update
apt-get install -y cri-o
systemctl start crio.service
sysctl -w net.ipv4.ip_forward=1
sed -i 's/#net.ipv4.ip_forward=1/net.ipv4.ip_forward=1/' /etc/sysctl.conf # persist after reboot

# criu
wget https://github.com/checkpoint-restore/criu/archive/refs/tags/v3.19.tar.gz
tar -xzvf v3.19.tar.gz
export DEBIAN_FRONTEND=noninteractive
apt install -y build-essential asciidoctor libprotobuf-dev
apt install -y libprotobuf-c-dev protobuf-c-compiler protobuf-compiler
apt install -y python3-protobuf pkg-config libbsd-dev
apt install -y iproute2 libnftables-dev libgnutls28-dev
apt install -y libnl-3-dev libnet-dev libcap-dev
cd criu-3.19
make install

# go
wget https://go.dev/dl/go1.22.3.linux-arm64.tar.gz
tar -C /usr/local -xzf go1.22.3.linux-arm64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile

ip netns add t0
ip netns set t0 10
ip link add vetht0a type veth peer name vetht0b
ip link set vetht0b netns t0
ip addr add 10.0.0.2/24 dev vetht0a
ip link set dev vetht0a up
ip netns exec t0 ip addr add 10.0.0.1/24 dev vetht0b
ip netns exec t0 ip link set dev vetht0b up

ip netns add t1
ip netns set t1 11
ip link add vetht1a type veth peer name vetht1b
ip link set  vetht1b netns t1
ip addr add 10.0.0.3/24 dev vetht1a
ip link set dev vetht1a up
ip netns exec t1 ip addr add 10.0.0.1/24 dev vetht1b
ip netns exec t1 ip link set dev vetht1b up
