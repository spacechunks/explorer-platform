/*
Explorer Platform, a platform for hosting and discovering Minecraft servers.
Copyright (C) 2024 Yannic Rieger <oss@76k.io>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define ETH_P_IP 0x0800
#define ETH_ALEN 6
#define TC_ACT_OK 0
#define IPPROTO_ICMP 1

/*
struct if_data {
    int if_idx;
    u8 mac_addr[ETH_ALEN];
};

struct {
    __unit(type, BPF_MAP_TYPE_HASH);
    __type(key, int);
    __type(value, struct if_data);
    __unit(max_entries, 256) // TODO: determine value
} port_to_if_data SEC(".maps");*/

struct cur {
        void *pos;
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

static __always_inline int parse_iphdr(void **data, void *data_end, struct iphdr **iph)
{
    if (*data + sizeof(struct iphdr) > data_end) {
        bpf_printk("2");
        return -1;
    }
    struct iphdr *tmp = *data;
    int len = tmp->ihl * 4;
    if (*data + len > data_end) {
        bpf_printk("1");
        return -1;
    }
    *iph = tmp;
    *data += len;
    return 0;
}

SEC("tc")
int ingress(struct __sk_buff *ctx)
{
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    if (ctx->protocol != bpf_htons(ETH_P_IP)) {
            return TC_ACT_OK;
    }

    struct ethhdr *ethh;
    if (parse_ethhdr(&data, data_end, &ethh))
            return TC_ACT_OK;

    struct iphdr *iph;
    if(parse_iphdr(&data, data_end, &iph))
            return TC_ACT_OK;

    if (iph->protocol != IPPROTO_ICMP) {
        return TC_ACT_OK;
    }

    bpf_printk("hd; %d", iph->protocol);
    return 0;
}

char _license[] SEC("license") = "GPL";
