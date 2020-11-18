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

#### nhost login

Login to your Nhost account

```
USAGE
  $ nhost login

DESCRIPTION
  ...
  Login to your Nhost account
```

#### nhost init

Initialize Nhost project

```
USAGE
  $ nhost init

DESCRIPTION
  ...
  Initializes the current working directory as a Nhost project
```

#### nhost dev

Start Nhost project for local development

```
USAGE
  $ nhost dev

DESCRIPTION
  ...
  Start the Nhost project with PostgreSQL, Hasura GraphQL Engine and Hasura Backend Plus (HBP)
```

#### nhost deploy

Deploy local migrations and metadata changes to Nhost production

```
USAGE
  $ nhost deploy

DESCRIPTION
  ...
  Deploy local migrations and metadata changes to Nhost production
```

#### nhost logout

Logout from your Nhost account

```
USAGE
  $ nhost logout

DESCRIPTION
  ...
  Logout from your Nhost account
```

### External dependencies

https://github.com/hasura/graphql-engine/tree/master/cli#installation

### Documentation

https://docs.nhost.io/cli