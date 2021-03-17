const { Command } = require("@oclif/command");
const chalk = require("chalk");
const util = require("util");
const exec = util.promisify(require("child_process").exec);

class DownCommand extends Command {
  async run() {
    this.log(`\n${chalk.white("bbb...")}`);

    const projectConfig = yaml.safeLoad(
      await readFile(`.nhost/nhost.yaml`, { encoding: "utf8" })
    );
    const projectId = projectConfig.project_id;

    // get env vars from remote
    const userData = await validateAuth();

    const projects = [
      ...userData.user.projects,
      ...userData.user.teams.flatMap(({ team }) => team.projects),
    ];

    const project = projects.find((project) => project.id === projectId);

    const envFile = `.env.development`;
  }
}

DownCommand.description = `bbb :D`;

module.exports = DownCommand;
