const { Command } = require("@oclif/command");
const chalk = require("chalk");
const util = require("util");
const exec = util.promisify(require("child_process").exec);

class DownCommand extends Command {
  async run() {
    await exec(
      "docker rm -f nhost_postgres nhost_hasura nhost_hbp nhost_minio nhost_api"
    );
    this.log(`\n${chalk.white("All Nhost services down")}`);
  }
}

DownCommand.description = `Stop and remove local Nhost containers started by \`nhost-dev\`
...
Stop and remove local Nhost containers started by \`nhost-dev\`
`;

module.exports = DownCommand;
