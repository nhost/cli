const { Command } = require("@oclif/command");
const yaml = require("js-yaml");
const util = require("util");
const fs = require("fs");
const { cli } = require("cli-ux");
const readFile = util.promisify(fs.readFile);
const { validateAuth } = require("../../util/login");

class DownCommand extends Command {
  async run() {
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
    console.log(`Environment variables for ${project.name}.\n`);

    cli.table(
      project.project_env_vars,
      {
        name: {
          header: "name",
          minWidth: 20,
        },
        dev_value: {
          header: "value (dev)",
        },
      },
      {
        printLine: this.log,
      }
    );
  }
  async catch(error) {
    console.log("");
    console.log(error);
  }
}

DownCommand.description = `List environment variables`;

module.exports = DownCommand;
