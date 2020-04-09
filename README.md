Nhost CLI
=========

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

Initialize Nhost project

```
USAGE
  $ nhost init

OPTIONS
  -d, --directory=directory  Where to create a project (working directory assumed if not specified)

DESCRIPTION
  ...
  Initializes a new project (or an existing one) with configuration for running the Nhost environment
```

### nhost dev

Start the project development environment

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

### Hasura GraphQL Engine CLI

https://github.com/hasura/graphql-engine/tree/master/cli#installation
