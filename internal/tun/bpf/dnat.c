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

#define TCP_DPORT_OFF (ETH_HLEN + sizeof(struct iphdr) + offsetof(struct tcphdr, dest))
#define IP_DST_OFF (ETH_HLEN + offsetof(struct iphdr, daddr))
#define BPF_F_PSEUDO_HDR 0x10

SEC("tc")
int dnat(struct __sk_buff *ctx)
{
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    if (ctx->protocol != bpf_htons(ETH_P_IP))
        return TC_ACT_OK;

    __be16 port;
    bpf_skb_load_bytes(ctx, TCP_DPORT_OFF, &port, sizeof(__be16));

    if (bpf_htons(port) != 25565) {
        return TC_ACT_OK;
    }

    /* d6:35:fc:1e:56:15 */
    u8 dest_mac[] = {
        214, 53, 252, 30, 86, 21
    };

   struct ethhdr *ethh;
   if (data + sizeof(struct ethhdr) > data_end) {
        return TC_ACT_OK;
   }

    ethh = data;

    __builtin_memcpy(ethh->h_dest, dest_mac, ETH_ALEN);
    __be32 dst = bpf_htonl(0xa000001);
    __u32 prev_dst;

    bpf_skb_load_bytes(ctx, IP_DST_OFF, &prev_dst, 4);
    bpf_skb_store_bytes(ctx, IP_DST_OFF, &new_ip, sizeof(dst), 0);
    bpf_l3_csum_replace(ctx, IP_CSUM_OFF, prev_dst, dst, sizeof(dst));
    bpf_l4_csum_replace(ctx, TCP_CSUM_OFF, prev_dst, dst,  BPF_F_PSEUDO_HDR | sizeof(dst));

    return bpf_redirect_peer(4, 0);
}

char _license[] SEC("license") = "GPL";
