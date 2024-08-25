#!/bin/bash

# Explorer Platform, a platform for hosting and discovering Minecraft servers.
# Copyright (C) 2024 Yannic Rieger <oss@76k.io>
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

apt update
apt-get install -y gnupg2 git

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
cd -

# go
wget https://go.dev/dl/go1.22.3.linux-arm64.tar.gz
tar -C /usr/local -xzf go1.22.3.linux-arm64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile

# cni plugins
git clone https://github.com/containernetworking/plugins.git
cd plugins
./build_linux.sh
cd -
mkdir -p /opt/cni
cp -r plugins/bin /opt/cni
ls /opt/cni
cp ptpnat /opt/cni/bin/ptpnat

# install ptpnat
# TODO: copy ptpnat binray
cp /root/10-ptpnat.conflist /top

# crictl
VERSION=v1.30.1 # check latest version in /releases page
ARCH=arm64
wget https://github.com/kubernetes-sigs/cri-tools/releases/download/$VERSION/crictl-$VERSION-linux-$ARCH.tar.gz
sudo tar zxvf crictl-$VERSION-linux-$ARCH.tar.gz -C /usr/local/bin
rm -f crictl-$VERSION-linux-$ARCH.tar.gz

# run nginx pod
crictl pull docker.io/nginx:stable-alpine-slim
pod=$(crictl -t 5m runp pod.json)
ctr=$(crictl -t 5m create $pod ctr.json pod.json)
crictl -t 5m start $ctr
