// const { Command, flags } = require("@oclif/command");
const { Command, flags } = require("@oclif/command");
const fs = require("fs");
const chalk = require("chalk");
const util = require("util");
const exec = util.promisify(require("child_process").exec);
const writeFile = util.promisify(fs.writeFile);
const nunjucks = require("nunjucks");

const spinnerWith = require("../util/spinner");
const selectProject = require("../util/projects");
const {
  authFileExists,
  readAuthFile,
  getCustomApiEndpoint,
  getNhostConfigTemplate,
} = require("../util/config");
const { validateAuth } = require("../util/login");

class InitCommand extends Command {
  projectOnHBPV2(project) {
    return project.backend_version.includes("v2");
  }

  async run() {
    const apiUrl = getCustomApiEndpoint();

    // check if hasura is installed
    try {
      await exec("command -v hasura");
    } catch {
      this.log(
        `${chalk.red(
          "Hasura CLI is missing!"
        )} Please follow the instructions here https://hasura.io/docs/1.0/graphql/manual/hasura-cli/install-hasura-cli.html`
      );
      this.exit(1);
    }

    // check if auth file exists
    if (!await authFileExists()) {
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

    let selectedProjectId;
    try {
      selectedProjectId = await selectProject(userData.user.projects);
    } catch (err) {
      this.log(`${chalk.red("Error!")} ${err.message}`);
      this.exit(1);
    }

    const project = userData.user.projects.find(
      (project) => project.id === selectedProjectId
    );
    
    // assume current working directory
    const directory = ".";

    // config.yaml holds configuration for GraphQL engine, PostgreSQL and HBP
    // it is also a requirement for hasura to work
    await writeFile(
      `${directory}/config.yaml`,
      nunjucks.renderString(getNhostConfigTemplate(), project)
    );

    // create a migrations directory if not present
    const migrationDirectory = `${directory}/migrations`;
    if (!fs.existsSync(migrationDirectory)) {
      fs.mkdirSync(migrationDirectory);
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

    let command = `hasura migrate create "init" --from-server --endpoint ${hasuraEndpoint} --schema "public" --schema "auth"`;
    if (adminSecret) {
      command += ` --admin-secret ${adminSecret}`;
    }

    let {spinner, stopSpinner} = spinnerWith(`Initializing ${project.name}`);
    try {
      await exec(command);

      // mark this migration as applied on the remote server
      // so that it doesn't get run there when promoting local
      // changes to that environment (redundant)
      const initMigration = fs.readdirSync("./migrations")[0];
      const version = initMigration.match(/^[0-9]+/)[0];
      command = `hasura migrate apply --version "${version}" --skip-execution --endpoint ${hasuraEndpoint}`;
      if (adminSecret) {
        command += ` --admin-secret ${adminSecret};`;
      }
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
