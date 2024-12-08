# datapath

This is the implementation of the datapath of the platform stack. For each Minecraft server container a unique port on the host 
system will be allocated. 

**Basic overview of functionality by interface**

* internet-facing interface
  * Packets arriving with a matching destination port will be redirected
    to the container-side veth peer of the corresponding server container.
* host veth peer
  * Packets arriving with a source port of `25565` will be redirected to the 
    internet-facing inetface.
  * arp requests 
* container veth peer
  * Packets leaving that have a destination port of `80` will be redirected to the HTTP
    listener of the transparent proxy.
  * Packets leaving that have a destination port of `53` will be redirected to
    the transparent proxy where they will be forwarded to a DNS server.
  * Packets leaving that don't have a destination port of `53`, `80` or `25565` will be
    redirected to the generic tcp listener of the transparent proxy.

**Systemd notes**

When using systemd, `MacAddressPolicy` needs to be set to `none`. Otherwise, there appears to be a race condition where our configured
MAC address configured on the host-side veth peer will not be picked up. This is because since version 242, systemd will set a persistent 
mac address on virtual interfaces. For more info see [here](https://lore.kernel.org/netdev/CAHXsExy+zm+twpC9Qrs9myBre+5s_ApGzOYU45Pt=sw-FyOn1w@mail.gmail.com/). 

```
# /etc/systemd/network/10-ignore.link
[Match]
OriginalName=*

[Link]
MACAddressPolicy=none
```

## bpf

Here is a quick explanation of what the purpose of our eBPF programs is.

**`dnat.c`**
Match incoming TCP packets by port and change their destination IP address and redirect them to the corresponding veth peer. 
This program will be run on main on the main physical interface, specifically TC ingress.

**`snat.c`**
Redirect arriving packets to internet-facing interface by changing source IP address.
This program will be run on the host-side veth peer, specifically TC ingress.

**`arp.c`**
Respond to ARP requests coming from the pod-side veth peer with the host-side peers MAC address.
This program will be run on the host-side veth peer, specifically TC ingress.

**includes**
* `include/bpf`
  * libbpf headers copied using the latest stable version.
* `include/linux`
  * `vmlinux.h` is missing defines like `ETH_ALEN` and such, so we copied those directly from the Linux kernel.
     files are stripped of any duplicate struct definitions.

Having those directly in our repo ensures builds are not machine-dependent.

