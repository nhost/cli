#!/bin/sh
set -e

arch="$(uname -m)"

docker build --build-arg HASURA_CLI_ARCH="$arch" -t nhost/hasura:v2.8.1 .
