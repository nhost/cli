const fs = require("fs");
const util = require("util")
const exec = util.promisify(require("child_process").exec);
const chalk = require("chalk");

async function checkForHasura() {
  try {
    await exec("command -v hasura");
  } catch {
    throw new Error(
      `${chalk.red(
        "\nHasura CLI is missing!"
      )} Please follow the instructions here https://hasura.io/docs/1.0/graphql/manual/hasura-cli/install-hasura-cli.html`
    );
  }
}

module.exports = checkForHasura;
