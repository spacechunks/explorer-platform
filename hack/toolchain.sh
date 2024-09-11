#!/bin/bash
GO_VERSION="1.23.1"
ARCH="arm64"
LLVM="18"

sudo apt update
sudo apt install make
curl https://apt.llvm.org/llvm.sh > llvm
sudo bash llvm $LLVM all
rm llvm

sudo wget https://go.dev/dl/go$GO_VERSION.linux-$ARCH.tar.gz
sudo tar -C /usr/local -xzf go$GO_VERSION.linux-$ARCH.tar.gz
rm go$GO_VERSION.linux-$ARCH.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile
