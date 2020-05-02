Nhost CLI
=========

*Only \*nix systems are currently supported. Windows support is coming soon!*


The easiast way to develop with Nhost in your development environment.

### Install

```bash
npm install -g nhost
```

or

```bash
$ git clone https://github.com/nhost/cli
$ cd cli
$ npm install
$ npm link
```

### Usage

### nhost init

Initialise Nhost project

```
USAGE
  $ nhost init

OPTIONS
  -d, --directory=directory  Where to create a project (working directory assumed if not specified)
  -e, --endpoint=https://hasura-a50fsaz6.nhost.app Endpoint of your GraphQL engine running on Nhost (Used when initialising from an existing project)
  -a, --admin-secret=secret GraphQl engine admin secret (if any)

DESCRIPTION
  ...
  Initialises a new project from scratch (or from an existing one) 
  Creates Configuration for running a Nhost environment
```

### nhost dev

Start development environment

```
USAGE
  $ nhost dev

DESCRIPTION
  ...
  Starts a complete Nhost environment with PostgreSQL, Hasura GraphQL Engine and Hasura Backend Plus (HBP)
```

### nhost help [COMMAND]

Display help for nhost

```
USAGE
  $ nhost help [COMMAND]

ARGUMENTS
  COMMAND  command to show help for

OPTIONS
  --all  see all commands in CLI
```

### External dependencies

https://github.com/hasura/graphql-engine/tree/master/cli#installation
