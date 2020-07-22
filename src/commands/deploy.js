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

class DeployCommand extends Command {
  async run() {
    const apiUrl = getCustomApiEndpoint();
    try {
      await checkForHasura();
    } catch (err) {
      this.log(err.message);
      this.exit(1);
    }

    // check if auth file exists
    if (!(await authFileExists())) {
      this.log(
        `${chalk.red(
          "No credentials found!"
        )} Please login first with ${chalk.bold.underline("nhost login")}`
      );
      this.exit(1);
    }

    // get auth config
    const auth = readAuthFile();
    let userData;
    try {
      userData = await validateAuth(apiUrl, auth);
    } catch (err) {
      this.log(`${chalk.red("Error!")} ${err.message}`);
      this.exit(1);
    }

    let { spinner, stopSpinner } = spinnerWith("deploying migrations");

    const projectConfig = yaml.safeLoad(
      await readFile("./.nhost/nhost.yaml", { encoding: "utf8" })
    );
    const projectID = projectConfig.project_id;

    const project = userData.user.projects.find(
      (project) => project.id === projectID
    );

    const hasuraEndpoint = `https://${project.project_domain.hasura_domain}`;
    const adminSecret = project.hasura_gqe_admin_secret;
    try {
      const { stdout } = await exec(
        `hasura migrate apply --endpoint=${hasuraEndpoint} --admin-secret=${adminSecret}`,
      );

      // TODO find out a better way of doing this
      const logLine = stdout.split("\n")[1];
      if (logLine && logLine.includes("nothing to apply")) {
        spinner.succeed("nothing to apply");
      } else { 
        spinner.succeed("migrations applied");
      }
      stopSpinner();
    } catch (err) {
      this.log(`${chalk.red("Error!")} ${err.message}`);
      this.exit(1);
    }
  }
}

DeployCommand.description = `Deploy local migrations to Nhost production
...
Deploy local migrations to Nhost production
`;

module.exports = DeployCommand;
