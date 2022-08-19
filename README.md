<div align="center">
  <h1>Nhost CLI</h1>
</div>

<div align="center">

[![Release](https://github.com/nhost/cli/actions/workflows/release.yaml/badge.svg)](https://github.com/nhost/cli/actions/workflows/release.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nhost/cli)](https://goreportcard.com/report/github.com/nhost/cli)
<a href="https://twitter.com/nhost" target="_blank" rel="noopener noreferrer">
<img src="https://img.shields.io/twitter/follow/nhost?style=social" />
</a>

</div>

[Nhost](http://nhost.io) is an open source Firebase alternative with GraphQL.

The Nhost CLI is used to get a local environment for local development.

The CLI creates a local Nhost development environment:

- [Postgres Database](https://www.postgresql.org/)
- [Hasura's GraphQL Engine](https://github.com/hasura/graphql-engine)
- [Hasura Auth](https://github.com/nhost/hasura-auth)
- [Hasura Storage](https://github.com/nhost/hasura-storage)
- [Minio S3](https://github.com/minio/minio)
- [Nhost Serverless Functions](https://github.com/nhost/functions)

## Get Started

Install the Nhost CLI:

```
sudo curl -L https://raw.githubusercontent.com/nhost/cli/main/get.sh | bash
```

Initialize a project:

```
nhost init
```

Initialize a project with a remote project as a starting point:

```
nhost init --remote
```

Start the development environment:

```
nhost up
```

## Documentation

- [Get started with Nhost CLI (longer version)](https://docs.nhost.io/platform/overview/get-started-with-nhost-cli)
- [Nhost CLI](https://docs.nhost.io/platform/cli)
- [Reference](https://docs.nhost.io/reference/cli)

## Dependencies

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [crl](https://curl.se/)
- [Git](https://git-scm.com/downloads)

## Supported Platforms:

- MacOS
- Linux
- Windows WSL2
