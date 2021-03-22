const { Command } = require("@oclif/command");
const fs = require("fs");
const chalk = require("chalk");
const util = require("util");
const mkdir = util.promisify(fs.mkdir);
const exists = util.promisify(fs.exists);
const writeFile = util.promisify(fs.writeFile);

const selectProject = require("../util/projects");
const { readAuthFile, getCustomApiEndpoint } = require("../util/config");
const { validateAuth } = require("../util/login");

class LinkCommand extends Command {
  async run() {
    const workingDir = ".";
    const dotNhost = `${workingDir}/.nhost`;
    const apiUrl = getCustomApiEndpoint();
    const auth = readAuthFile();
    let userData;
    try {
      userData = await validateAuth(apiUrl, auth);
    } catch (err) {
      return this.log(`${chalk.red("Error!")} ${err.message}`);
    }

    // personal and team projects
    const projects = [
      ...userData.user.projects,
      ...userData.user.teams.flatMap(({ team }) => team.projects),
    ];

    if (projects.length === 0) {
      return this.log(
        `\nWe couldn't find any projects related to this account, go to ${chalk.bold.underline(
          "https://console.nhost.io/new"
        )} and create one`
      );
    }

    let selectedProjectId;
    try {
      selectedProjectId = await selectProject(projects);
    } catch (err) {
      return this.log(`${chalk.red("Error!")} ${err.message}`);
    }

    const project = projects.find(
      (project) => project.id === selectedProjectId
    );

    // .nhost is used for nhost specific configuration
    if (!(await exists(dotNhost))) {
      await mkdir(dotNhost);
    }
    await writeFile(
      `${dotNhost}/nhost.yaml`,
      `project_id: ${selectedProjectId}`
    );

    console.log(`Project linked: ${project.name}`);
  }
}

LinkCommand.description = `Link Nhost Project
...
Link Nhost project
`;

module.exports = LinkCommand;
