#pragma once

#define TC_ACT_UNSPEC (-1)
#define TC_ACT_OK 0
#define IP_ADDR_LEN 4
#define BPF_F_PSEUDO_HDR 0x10

#define ARPOP_REQUEST 1
#define ARPOP_REPLY 2

#define IP_CSUM_OFF (ETH_HLEN + offsetof(struct iphdr, check))
#define TCP_CSUM_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, check))

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

void __always_inline change_ip_addr(__be32 new, __u32 offset) {
    __be32 prev;
    bpf_skb_load_bytes(ctx, offset, &prev, IP_ADDR_LEN);
    bpf_skb_store_bytes(ctx, offset, &new, sizeof(new), 0);
    bpf_l3_csum_replace(ctx, IP_CSUM_OFF, prev, new, sizeof(new));
    bpf_l4_csum_replace(ctx, TCP_CSUM_OFF, prev, new, BPF_F_PSEUDO_HDR | sizeof(new));
}

void __always_inline change_port(__be16 new, __u32 offset) {
    __be16 prev;
    bpf_skb_load_bytes(ctx, offset, &prev, sizeof(__be16));
    bpf_skb_store_bytes(ctx, offset, new, sizeof(__be16))
    bpf_l4_csum_replace(ctx, TCP_CSUM_OFF, prev, new, 0);
}
