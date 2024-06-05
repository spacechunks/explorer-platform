#!/bin/bash
hcloud server create \
  --name dev \
  --type cax21 \
  --image debian-12 \
  --ssh-key yannic-mac-work \

ip=$(hcloud server ip dev)
echo "$IP"
hcloud server ssh dev 'bash -s' < ./scripts/install-node.sh
