const { Command } = require("@oclif/command");
const nunjucks = require("nunjucks");
const fs = require("fs");
const chalk = require("chalk");
const util = require("util");
const exec = util.promisify(require("child_process").exec);
const exists = util.promisify(fs.exists);
const writeFile = util.promisify(fs.writeFile);
const mkdir = util.promisify(fs.mkdir);

const spinnerWith = require("../util/spinner");
const selectProject = require("../util/projects");
const {
  authFileExists,
  readAuthFile,
  getCustomApiEndpoint,
  getNhostConfigTemplate,
} = require("../util/config");
const { validateAuth } = require("../util/login");
const checkForHasura = require("../util/dependencies");

class InitCommand extends Command {
  projectOnHBPV2(project) {
    return project.backend_version.includes("v2");
  }

  async run() {
    const apiUrl = getCustomApiEndpoint();
    // assume current working directory
    const directory = ".";

    // check if hasura is installed
    try {
      await checkForHasura();
    } catch (err) {
      console.log(err.message);
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

    // check if project is already initialized
    if (await exists(`${directory}/config.yaml`)) {
      this.log(
        `\n${chalk.white(
          "This directory seems to have a project already configured, skipping"
        )}`
      );
      this.exit();
    }

    // personal projects + projects from teams the user is a member of
    const projects = [
      ...userData.user.projects,
      ...userData.user.teams.flatMap(({team}) => team.projects),
    ];

    if (projects.length === 0) {
      this.log(
        `\nWe couldn't find any projects related to this account, go to ${chalk.bold.underline(
          "https://console.nhost.io/new-project"
        )} and create one`
      );
      this.exit();
    }

    let selectedProjectId;
    try {
      selectedProjectId = await selectProject(projects);
    } catch (err) {
      this.log(`${chalk.red("Error!")} ${err.message}`);
      this.exit(1);
    }

    const project = projects.find(
      (project) => project.id === selectedProjectId
    );

    // .nhost is used for nhost specific configuration
    const dotNhost = `${directory}/.nhost`;
    await mkdir(dotNhost);
    await writeFile(
      `${dotNhost}/nhost.yaml`,
      `project_id: ${selectedProjectId}`
    );

    // config.yaml holds configuration for GraphQL engine, PostgreSQL and HBP
    // it is also a requirement for hasura to work
    await writeFile(
      `${directory}/config.yaml`,
      nunjucks.renderString(getNhostConfigTemplate(), project)
    );

    // create directory for migrations
    const migrationDirectory = `${directory}/migrations`;
    if (!fs.existsSync(migrationDirectory)) {
      fs.mkdirSync(migrationDirectory);
    }

    // create directory for metadata
    const metadataDirectory = `${directory}/metadata`;
    if (!fs.existsSync(metadataDirectory)) {
      fs.mkdirSync(metadataDirectory);
    }
    // create or append to .gitignore
    const ignoreFile = `${directory}/.gitignore`;
    fs.writeFileSync(ignoreFile, "\nconfig.yaml\n.nhost\ndb_data\nminio_data", {
      flag: "a",
    });

    // .env.development for hasura webhooks, headers, etc
    const envFile = `${directory}/.env.development`;
    if (!fs.existsSync(envFile)) {
      await writeFile(envFile, "# webhooks and headers\n");
    }

    const hasuraEndpoint = `https://${project.project_domain.hasura_domain}`;
    const adminSecret = project.hasura_gqe_admin_secret;

    let { spinner, stopSpinner } = spinnerWith(`Initializing ${project.name}`);

    const commonOptions = `--endpoint ${hasuraEndpoint} --admin-secret ${adminSecret} --skip-update-check`;
    try {
      // create migrations from remote 
      let command = `hasura migrate create "init" --from-server --schema "public" --schema "auth" ${commonOptions}`;
      await exec(command);

      // mark this migration as applied (--skip-execution) on the remote server
      // so that it doesn't get run there when promoting local
      // changes to that environment 
      const initMigration = fs.readdirSync("./migrations")[0];
      const version = initMigration.match(/^[0-9]+/)[0];
      command = `hasura migrate apply --version "${version}" --skip-execution --endpoint ${hasuraEndpoint} --admin-secret ${adminSecret}`;
      await exec(command);

      // create metadata from remote
      command = `hasura metadata export ${commonOptions}`;
      await exec(command);

      //  create seeds from remote
      command = `hasura seeds create roles_and_providers --from-table auth.roles --from-table auth.providers ${commonOptions}`;
      await exec(command);

      // TODO: rethink the necessity of citext
      // prepend the contents of the sql file with the installation of citext
      // this is a requirement for HBPv2
      if (this.projectOnHBPV2(project)) {
        const sqlPath = `./migrations/${initMigration}/up.sql`;
        const data = fs.readFileSync(sqlPath);
        const sql = fs.openSync(sqlPath, "w+");
        const citext = Buffer.from("CREATE EXTENSION IF NOT EXISTS citext;\n");
        fs.writeSync(sql, citext, 0, citext.length, 0);
        fs.writeSync(sql, data, 0, data.length, citext.length);
        fs.close(sql);
      }

      // write ENV variables to .env.development (webhooks and headers)
      await writeFile(
        `${directory}/.env.development`,
        project.hasura_gqe_custom_env_variables
          .map((envVar) => `${envVar.key}=${envVar.value}`)
          .join("\n"),
        { flag: "a" }
      );

      await writeFile(ignoreFile, "\n.env.development", { flag: "a" });
    } catch (error) {
      spinner.fail();
      stopSpinner();
      this.log(`${chalk.red("Error!")} ${error.message}`);
      this.exit(1);
    }

    spinner.succeed();
    stopSpinner();

    this.log(`${chalk.green("Nhost project successfully initialized")}`);
  }
}

InitCommand.description = `Initialize current working directory with Nhost project
...
Initialize current working directory with Nhost project 
`;

module.exports = InitCommand;
