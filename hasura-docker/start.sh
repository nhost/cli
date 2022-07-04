#!/bin/bash

# Workaround for https://github.com/hasura/graphql-engine/issues/2824#issuecomment-801293056
socat TCP-LISTEN:8080,fork TCP:graphql-engine:8080 &
socat TCP-LISTEN:9695,fork,reuseaddr,bind=hasura-console TCP:127.0.0.1:9695 &
socat TCP-LISTEN:9693,fork,reuseaddr,bind=hasura-console TCP:127.0.0.1:9693 &

# Run console if specified
if [[ -v HASURA_RUN_CONSOLE ]]; then
    echo "Starting console..."
    hasura console --skip-update-check --log-level DEBUG --address "127.0.0.1" --no-browser || exit 1
else
    sleep infinity
fi
