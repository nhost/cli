nhost
=========

Nhost&#39;s CLI. Get a local Nhost backend for development.

[![oclif](https://img.shields.io/badge/cli-oclif-brightgreen.svg)](https://oclif.io)

<!-- toc -->
* [Quick Start](#quick-start)
* [Usage](#usage)
* [Commands](#commands)
<!-- tocstop -->

# Quick Start
<!-- quickstart-->

### Install

```bash
$ git clone https://github.com/nhost/cli
$ cd cli
$ npm install
$ npm link
```

### Usage

Open a new terminal window

```bash
$ nhost init -d facebook2
$ cd facebook2
$ nhost dev
```

<!-- quickstartstop -->



# Usage
<!-- usage -->
```sh-session
$ npm install -g nhost
$ nhost COMMAND
running command...
$ nhost (-v|--version|version)
nhost/0.0.1 darwin-x64 node-v12.13.1
$ nhost --help [COMMAND]
USAGE
  $ nhost COMMAND
...
```
<!-- usagestop -->
# Commands
<!-- commands -->
* [`nhost init`](#nhost-init)
* [`nhost dev`](#nhost-dev)
* [`nhost help [COMMAND]`](#nhost-help-command)

## `nhost dev`

Describe the command here

## `nhost init`

Describe the command here

```
USAGE
  $ nhost init

OPTIONS
  -d, --directory=directory  directory where to create the files

DESCRIPTION
  ...
  Extra documentation goes here
```

_See code: [src/commands/init.js](https://github.com/nhost/cli/blob/v0.0.1/src/commands/init.js)_

```
USAGE
  $ nhost dev

OPTIONS
  -n, --name=name  name to print

DESCRIPTION
  ...
  Extra documentation goes here
```

_See code: [src/commands/dev.js](https://github.com/nhost/cli/blob/v0.0.1/src/commands/dev.js)_

## `nhost help [COMMAND]`

display help for nhost

```
USAGE
  $ nhost help [COMMAND]

ARGUMENTS
  COMMAND  command to show help for

OPTIONS
  --all  see all commands in CLI
```

_See code: [@oclif/plugin-help](https://github.com/oclif/plugin-help/blob/v2.2.3/src/commands/help.ts)_

<!-- commandsstop -->
