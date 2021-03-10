const { Command } = require("@oclif/command");
const yaml = require("js-yaml");
const checkForHasura = require("../util/dependencies");
const spinnerWith = require("../util/spinner");
const {
  authFileExists,
  readAuthFile,
  getCustomApiEndpoint,
} = require("../util/config");
const { validateAuth } = require("../util/login");
const chalk = require("chalk");

const fs = require("fs");
const util = require("util");
const exec = util.promisify(require("child_process").exec);
const readFile = util.promisify(fs.readFile);
const exists = util.promisify(fs.exists);

class DeployCommand extends Command {
  async run() {
    const workingDir = ".";
    const nhostDir = `${workingDir}/nhost`;
    const dotNhost = `${workingDir}/.nhost`;

    const apiUrl = getCustomApiEndpoint();

    try {
      await checkForHasura();
    } catch (err) {
      return this.log(err.message);
    }

    // check if auth file exists
    if (!(await authFileExists())) {
      return this.log(
        `${chalk.red(
          "No credentials found!"
        )} Please login first with ${chalk.bold.underline("nhost login")}`
      );
    }

    // get auth config
    const auth = readAuthFile();
    let userData;
    try {
      userData = await validateAuth(apiUrl, auth);
    } catch (err) {
      return this.log(`${chalk.red("Error!")} ${err.message}`);
    }

    if (!(await exists(`${dotNhost}`))) {
      return this.log(
        `${chalk.red(
          "Error!"
        )} this directory doesn't seem to be a valid project, please run ${chalk.underline.bold(
          "nhost init"
        )} to initialize it`
      );
    }

    const projectConfig = yaml.safeLoad(
      await readFile(`${dotNhost}/nhost.yaml`, { encoding: "utf8" })
    );
    const projectID = projectConfig.project_id;

    const projects = [
      ...userData.user.projects,
      ...userData.user.teams.flatMap(({ team }) => team.projects),
    ];
    const project = projects.find((project) => project.id === projectID);

    if (!project) {
      return this.log(
        `${chalk.red("Error!")} we couldn't find this project in our system`
      );
    }

    const hasuraEndpoint = `https://${project.project_domain.hasura_domain}`;
    const adminSecret = project.hasura_gqe_admin_secret;
    try {
      let { spinner } = spinnerWith("deploying migrations");
      await exec(
        `hasura migrate apply --endpoint=${hasuraEndpoint} --admin-secret=${adminSecret}`,
        { cwd: nhostDir }
      );
      spinner.succeed("migrations deployed");

      ({ spinner } = spinnerWith("deploying metadata"));
      await exec(
        `hasura metadata apply --endpoint=${hasuraEndpoint} --admin-secret=${adminSecret}`,
        { cwd: nhostDir }
      );
      spinner.succeed("metadata deployed");
    } catch (err) {
      return this.log(`\n${chalk.red("Error!")} ${err.message}`);
    }
  }
}

DeployCommand.description = `Deploy local migrations and metadata changes to Nhost production
...
Deploy local migrations and metadata changes to Nhost production
`;

module.exports = DeployCommand;
