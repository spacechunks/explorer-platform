MAJOR_VERSION=1.30
curl -fsSL https://pkgs.k8s.io/addons:/cri-o:/stable:/v$MAJOR_VERSION/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/cri-o-apt-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/cri-o-apt-keyring.gpg] https://pkgs.k8s.io/addons:/cri-o:/stable:/v$MAJOR_VERSION/deb/ /" | tee /etc/apt/sources.list.d/cri-o.list
apt-get update
apt-get install -y cri-o
systemctl start crio.service
sysctl -w net.ipv4.ip_forward=1
sed -i 's/#net.ipv4.ip_forward=1/net.ipv4.ip_forward=1/' /etc/sysctl.conf # persist after reboot

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
