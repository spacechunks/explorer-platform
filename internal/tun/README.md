# tun

This is the CNI component of the platform stack. It tunnels user traffic to running containers using eBPF snat and dnat.

## bpf

Here is a quick explanation of what the purpose of our eBPF programs is.

**`snat.c`**
Match incoming TCP packets by port and change their source IP address and redirect them to the corresponding veth peer. 
This program will be run on main on the main physical interface, specifically TC ingress.

**`dnat.c`**
Redirect arriving packets to main physical interface by changing destination IP address.
This program will be run on the host-side veth peer, specifically TC ingress.

**includes**
* `include/bpf`
  * libbpf headers copied using the latest stable version.
* `include/linux`
  * `vmlinux.h` is missing defines like `ETH_ALEN` and such, so we copied those directly from the Linux kernel.
     files are stripped of any duplicate struct definitions.

Having those directly in our repo ensures builds are not machine-dependent.
