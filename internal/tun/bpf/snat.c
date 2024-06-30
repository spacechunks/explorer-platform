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

#include "types.h"
#include "bpf/bpf_helpers.h"
#include "bpf/bpf_endian.h"
#include "linux/if_ether.h"
#include "vmlinux.h"
#include "net_helpers.h"

#define IP_SRC_OFF (ETH_HLEN + offsetof(struct iphdr, saddr))

struct snat_config_entry {
    __le32 ip_addr;
    __u8 iface_idx;
};

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __le16); /* host network order */
    __type(value, struct snat_config_entry);
    __uint(max_entries, 1);
} snat_config SEC(".maps");

SEC("tc")
int snat(struct __sk_buff *ctx)
{
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    if (ctx->protocol != bpf_htons(ETH_P_IP))
        return TC_ACT_OK;

    __u8 idx = 0;
    struct snat_config_entry *entry = bpf_map_lookup_elem(&snat_config, &idx);
    if (entry == NULL) {
        bpf_printk("no snat config entry found");
        return TC_ACT_OK;
    }

    __be32 src = bpf_htonl(entry->ip_addr);
    __be32 prev_src;

    bpf_skb_load_bytes(ctx, IP_SRC_OFF, &prev_src, 4);
    bpf_skb_store_bytes(ctx, IP_SRC_OFF, &src, sizeof(src), 0);
    bpf_l3_csum_replace(ctx, IP_CSUM_OFF, prev_src, src, sizeof(src));
    bpf_l4_csum_replace(ctx, TCP_CSUM_OFF, prev_src, src,  BPF_F_PSEUDO_HDR | sizeof(src));

    /* this fills in the l2 address for us */
    return bpf_redirect_neigh(entry->iface_idx, NULL, 0, 0);
}

char _license[] SEC("license") = "GPL";
