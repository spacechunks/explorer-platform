#!/bin/bash
docker build -t unpack-img -f Dockerfile.unpack .
docker image save unpack-img > unpack-img.tar.gz

docker build -t repack-img -f Dockerfile.repack .
docker image save repack-img > repack-img.tar.gz
