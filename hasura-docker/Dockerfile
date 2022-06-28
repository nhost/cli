FROM debian:buster-slim

ARG HASURA_VERSION=2.8.1

WORKDIR /usr/src/hasura

RUN apt-get update -y && apt-get install -y curl socat

RUN curl -L -o /usr/local/bin/hasura https://github.com/hasura/graphql-engine/releases/download/v${HASURA_VERSION}/cli-hasura-linux-arm64
RUN chmod +x /usr/local/bin/hasura

COPY start.sh /usr/local/bin/hasura-console-start

RUN chmod +x /usr/local/bin/hasura-console-start

ENTRYPOINT ["bash", "-c", "hasura-console-start"]
