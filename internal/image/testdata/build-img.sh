#!/bin/bash
docker build -t chunks-testimg .
docker image save chunks-testimg > img.tar.gz
