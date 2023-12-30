#!/bin/bash
docker build -t chunks-testimg ./internal/image/testdata
docker image save chunks-testimg > ./internal/image/testdata/img.tar.gz
