#!/bin/bash

hcloud server delete nodedev-yannic
hcloud server create --name nodedev-yannic --type cax21 --image debian-12 --ssh-key yannic-muc

ip=$(hcloud server ip nodedev-yannic)
sleep 30 # takes a bit of time until the server is reachable from the network
scp -r -o StrictHostKeyChecking=no nodedev/* root@$ip:/root
ssh -o StrictHostKeyChecking=no root@$ip '/root/provision.sh'
