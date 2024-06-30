#pragma once

#define TC_ACT_OK 0
#define IP_CSUM_OFF (ETH_HLEN + offsetof(struct iphdr, check))
#define TCP_CSUM_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, check))

static __always_inline int parse_ethhdr(void **data, void *data_end, struct ethhdr **ethh)
{
    int len = sizeof(struct ethhdr);
    if (*data + len > data_end)
        return -1;

    *ethh = *data;
    *data += len;
    return 0;
}
