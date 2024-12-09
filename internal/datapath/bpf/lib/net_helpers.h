#pragma once

#define TC_ACT_UNSPEC    (-1)
#define TC_ACT_SHOT      2
#define TC_ACT_OK        0
#define IP_ADDR_LEN      4
#define BPF_F_PSEUDO_HDR 0x10

#define ARPOP_REQUEST 1
#define ARPOP_REPLY   2

#define IP_CSUM_OFF   (ETH_HLEN + offsetof(struct iphdr, check))
#define IP_DST_OFF    (ETH_HLEN + offsetof(struct iphdr, daddr))
#define IP_SRC_OFF    (ETH_HLEN + offsetof(struct iphdr, saddr))
#define TCP_CSUM_OFF  (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, check))
#define TCP_DPORT_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, dest))
#define TCP_SPORT_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, source))
#define UDP_CSUM_OFF  (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct udphdr, check))
#define UDP_DPORT_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct udphdr, dest))
#define UDP_SPORT_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct udphdr, source))

#define MC_SERVER_PORT 25565

/*
 * the mac address configured on the host-side veth peer.
 * => 7e:90:c4:ed:df:d0
 */
const __u8 host_peer_mac[ETH_ALEN] = {
    126, 144, 196, 237, 223, 208
};

static __always_inline int parse_ethhdr(void **data, void *data_end, struct ethhdr **ethh)
{
    int len = sizeof(struct ethhdr);
    if (*data + len > data_end)
        return -1;

    *ethh = *data;
    *data += len;
    return 0;
}

static void __always_inline rewrite_ip_addr(struct __sk_buff **ctx, __be32 new, __u32 offset, __u8 proto)
{
    __be32 prev;
    bpf_skb_load_bytes(*ctx, offset, &prev, IP_ADDR_LEN);
    bpf_skb_store_bytes(*ctx, offset, &new, sizeof(new), 0);
    bpf_l3_csum_replace(*ctx, IP_CSUM_OFF, prev, new, sizeof(new));

    if (proto == IPPROTO_TCP)
        bpf_l4_csum_replace(*ctx, TCP_CSUM_OFF, prev, new, BPF_F_PSEUDO_HDR | sizeof(new));
    if (proto == IPPROTO_UDP)
        bpf_l4_csum_replace(*ctx, UDP_CSUM_OFF, prev, new, BPF_F_PSEUDO_HDR | sizeof(new));
}

static void __always_inline rewrite_port(struct __sk_buff **ctx, __be16 new, __u32 offset, __u8 proto)
{
    __be16 prev;
    bpf_skb_load_bytes(*ctx, offset, &prev, sizeof(prev));
    bpf_skb_store_bytes(*ctx, offset, &new, sizeof(new), 0);

    if (proto == IPPROTO_TCP) {
        bpf_l4_csum_replace(*ctx, TCP_CSUM_OFF, prev, new, 0);
        return;
    }

    if (proto == IPPROTO_UDP) {
        bpf_l4_csum_replace(*ctx, UDP_CSUM_OFF, prev, new, 0);
        return;
    }
}

static __be16 __always_inline get_port(struct __sk_buff **ctx, __u32 tcp_offset, __u32 udp_offset, __u8 proto)
{
    __be16 port;
    if (proto == IPPROTO_TCP) {
        bpf_skb_load_bytes(*ctx, tcp_offset, &port, sizeof(port));
        return port;
    }

    if (proto == IPPROTO_UDP) {
        bpf_skb_load_bytes(*ctx, udp_offset, &port, sizeof(port));
        return port;
    }

    return 0;
}
