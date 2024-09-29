#pragma once

#define TC_ACT_UNSPEC (-1)
#define TC_ACT_OK 0
#define IP_ADDR_LEN 4
#define BPF_F_PSEUDO_HDR 0x10

#define ARPOP_REQUEST 1
#define ARPOP_REPLY 2

#define IP_CSUM_OFF (ETH_HLEN + offsetof(struct iphdr, check))
#define TCP_CSUM_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, check))

/*
 * the mac address configured on the host-side veth.
 * => 7e:90:c4:ed:df:d0
 */
const __u8 host_veth_mac[ETH_ALEN] = {
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
