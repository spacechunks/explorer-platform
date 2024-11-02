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
#include "lib/net_helpers.h"

#define TPROXY_HTTP_PORT 9080
#define TPROXY_DNS_PORT  9053
#define TPROXY_TCP_PORT  9111

#define DPORT_DNS  53
#define DPORT_HTTP 80

#define SO_ORIGINAL_DST 80
#define AF_INET         2

struct original_dst_entry {
    __be32 ip_addr;
    __be16 port;
};

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH); /* use LRU so we dont fill up the map */
    __type(key, __u64);
    __type(value, struct original_dst_entry);
    __uint(max_entries, 10000);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} original_dst_map SEC(".maps");

SEC("tc")
int ctr_peer_egress(struct __sk_buff *ctx)
{
    // TODO: get ip address of current iface
    //       use map[ctx->ifindex]

    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    if (ctx->protocol != bpf_htons(ETH_P_IP))
        return TC_ACT_UNSPEC;

    struct iphdr *iph = data + ETH_HLEN;
    if (data + ETH_HLEN + sizeof(*iph) > data_end)
        return TC_ACT_OK;

    /*
     * save values of fields here, because later due to
     * changing the packet buffer we cannot access iph
     * directly without performing the bound checks again.
     */
    __u8 proto = iph->protocol;
    __be32 daddr = iph->daddr;

    __be16 dport = get_port(&ctx, TCP_DPORT_OFF, UDP_DPORT_OFF, proto);
    if (dport == 0) {
        bpf_printk("tproxy: dst port: invalid protocol %d", proto);
        return TC_ACT_OK;
    }

    __be16 sport = get_port(&ctx, TCP_SPORT_OFF, UDP_SPORT_OFF, proto);
    if (sport == 0) {
        bpf_printk("tproxy: src port: invalid protocol %d", proto);
        return TC_ACT_OK;
    }

    /* we only handle non minecraft-related traffic */
    if (sport == bpf_htons(MC_SERVER_PORT))
        return TC_ACT_OK;

    /*
     * use packed u64 as key for hash map, because when using a
     * struct as a key value things stopped working. bpf_map_lookup_elem in
     * host_peer_egress returned NULL consistently. interestingly only TCP
     * was affected. UDP worked as normal. another thing that was very
     * interesting to observe is that when running
     *
     *      ./pwru --filter-ifname vetht0a --output-tuple
     *
     * the first connection attempt failed, due to bpf_map_lookup_elem
     * in host_peer_egress returning NULL, but subsequent connections
     * succeeded. stopping pwru caused all connections to fail again
     * with the same error as previously.
     * pwru version used was v1.0.8.
     */
    __u64 key = (0 /*TODO: veth pair id*/ << 24) | (sport << 8) | proto;
    struct original_dst_entry val = {
        .port = dport,
        .ip_addr = daddr,
    };

    bpf_map_update_elem(&original_dst_map, &key, &val, BPF_ANY);

    __be32 dest = bpf_htonl(167772162); // 10.0.0.2
    rewrite_ip_addr(&ctx, dest, IP_DST_OFF, proto);

    if (dport == bpf_htons(DPORT_DNS)) {
        if (proto == IPPROTO_TCP)
            rewrite_port(&ctx, bpf_htons(TPROXY_DNS_PORT), TCP_DPORT_OFF, IPPROTO_TCP);
        if (proto == IPPROTO_UDP)
            rewrite_port(&ctx, bpf_htons(TPROXY_DNS_PORT), UDP_DPORT_OFF, IPPROTO_UDP);
        return TC_ACT_OK;
    }

    if (dport == bpf_htons(DPORT_HTTP)) {
        rewrite_port(&ctx, bpf_htons(TPROXY_HTTP_PORT), TCP_DPORT_OFF, IPPROTO_TCP);
        return TC_ACT_OK;
    }

    rewrite_port(&ctx, bpf_htons(TPROXY_TCP_PORT), TCP_DPORT_OFF, IPPROTO_TCP);
    return TC_ACT_OK;
}

SEC("tc")
int host_peer_egress(struct __sk_buff *ctx)
{
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    if (ctx->protocol != bpf_htons(ETH_P_IP))
        return TC_ACT_OK;

    struct iphdr *iph = data + ETH_HLEN;
    if (data + ETH_HLEN + sizeof(*iph) > data_end)
        return TC_ACT_OK;

    __u8 proto = iph->protocol;
    __be16 dport = get_port(&ctx, TCP_DPORT_OFF, UDP_DPORT_OFF, proto);

    __u64 key = (0 << 24) | (dport << 8) | proto;
    struct original_dst_entry *e = bpf_map_lookup_elem(&original_dst_map, &key);
    if (e == NULL) {
        bpf_printk("tproxy: egress: no entry for port %d/%d", proto, bpf_ntohs(dport));
        return TC_ACT_OK;
    }

    rewrite_ip_addr(&ctx, e->ip_addr, IP_SRC_OFF, proto);
    rewrite_port(&ctx, e->port, TCP_SPORT_OFF, UDP_SPORT_OFF);
    return TC_ACT_OK;
}

SEC("cgroup/getsockopt")
int getsockopt(struct bpf_sockopt *ctx)
{
    if (ctx->optname != SO_ORIGINAL_DST) return 1;
    /* only support ipv4 at the moment */
    if (ctx->sk->family != AF_INET) return 1;

    /* this is the port the packet was sent from */
    __be16 client_port = ctx->sk->dst_port;
    __u8 proto = ctx->sk->protocol;

     __u64 key = (0 << 24) | (client_port << 8) | proto;
    struct original_dst_entry *e = bpf_map_lookup_elem(&original_dst_map, &key);
    if (e == NULL) {
        bpf_printk("tproxy: getsockopt: no entry for port %d/%d", proto, bpf_ntohs(client_port));
        return 0;
    }

    struct sockaddr_in *sa = ctx->optval;
    if ((void*)(sa + 1) > ctx->optval_end) return 1;

    ctx->optlen = sizeof(*sa);
    sa->sin_family = ctx->sk->family;
    sa->sin_addr.s_addr = e->ip_addr;
    sa->sin_port = e->port;
    ctx->retval = 0;

    return 1;
}

char _license[] SEC("license") = "GPL";
