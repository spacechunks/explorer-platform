#!/usr/bin/bash
docker run --name xds --rm -d \
  --net=host \
  -v ./nodedev/envoy-xds.yaml:/etc/envoy/envoy.yaml \
  envoyproxy/envoy:v1.31-latest  \
  -c /etc/envoy/envoy.yaml \
  -l debug \
  && docker logs -f xds
