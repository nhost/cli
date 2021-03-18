const { Command } = require("@oclif/command");
const nunjucks = require("nunjucks");
const fs = require("fs");
const fetch = require("node-fetch");
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

  async _getExtensions(hasuraEndpoint, adminSecret) {
    const command = `curl -d '{"type": "run_sql", "args": {"sql": "SELECT * FROM pg_extension;"}}' -H 'X-Hasura-Admin-Secret: ${adminSecret}' ${hasuraEndpoint}/v1/query`;
    try {
      const response = await exec(command);
      const data = JSON.parse(response.stdout).result.splice(1); // remove head (first row)
      const extensions = data.map((row) => row[1]);
      return extensions;
    } catch (error) {
      console.log(error);
      console.log("Error getting extensions");
    }
  }

  async run() {
    const apiUrl = getCustomApiEndpoint();
    // assume current working directory
    const workingDir = ".";
    const nhostDir = `${workingDir}/nhost`;
    const dotNhost = `${workingDir}/.nhost`;

    // check if hasura is installed
    try {
      await checkForHasura();
    } catch (err) {
      return this.log(`${chalk.red("Error!")} ${err.message}`);
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

    // check if project is already initialized
    if (await exists(nhostDir)) {
      return this.log(
        `\n${chalk.white(
          "This directory seems to have a project already configured at ./nhost, skipping"
        )}`
      );
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

    const remoteHasuraVersion = project.hasura_gqe_version;
    const dockerImage = `nhost/hasura-cli-docker:${remoteHasuraVersion}`;

    // create root nhost folder
    await mkdir(nhostDir);
    // .nhost is used for nhost specific configuration
    await mkdir(dotNhost);
    await writeFile(
      `${dotNhost}/nhost.yaml`,
      `project_id: ${selectedProjectId}`
    );

    // config.yaml holds configuration for GraphQL engine, PostgreSQL and HBP
    // it is also a requirement for hasura to work
    await writeFile(
      `${nhostDir}/config.yaml`,
      nunjucks.renderString(getNhostConfigTemplate(), project)
    );

    // create directory for migrations
    const migrationDirectory = `${nhostDir}/migrations`;
    if (await !exists(migrationDirectory)) {
      await mkdir(migrationDirectory);
    }

    // create directory for metadata
    const metadataDirectory = `${nhostDir}/metadata`;
    if (await !exists(metadataDirectory)) {
      await mkdir(metadataDirectory);
    }
    // create or append to .gitignore
    const ignoreFile = `${workingDir}/.gitignore`;

    await writeFile(ignoreFile, ".nhost\napi/node_modules", {
      flag: "a",
    });

    // .env.development for hasura webhooks, headers, etc
    const envFile = `${workingDir}/.env.development`;
    if (await !exists(envFile)) {
      await writeFile(envFile, "# webhooks and headers\n");
    }

    const hasuraEndpoint = `https://${project.project_domain.hasura_domain}`;
    const adminSecret = project.hasura_gqe_admin_secret;

    let { spinner, stopSpinner } = spinnerWith(`Initializing ${project.name}`);

    try {
      // clear current migration information from remote
      await fetch(`${hasuraEndpoint}/v1/query`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-hasura-admin-secret": adminSecret,
        },
        body: JSON.stringify({
          type: "run_sql",
          args: {
            sql: "TRUNCATE hdb_catalog.schema_migrations;",
          },
        }),
      });

      const commonOptions = `--endpoint ${hasuraEndpoint} --admin-secret ${adminSecret} --skip-update-check`;

      // create migrations from remote
      // this migration will be auto applied
      let command = `docker run --rm -v $(pwd):/hasuracli ${dockerImage} migrate create "init" --from-server --schema "public" --schema "auth" ${commonOptions}`;
      await exec(command, { cwd: nhostDir });

      // // mark this migration as applied (--skip-execution) on the remote server
      // // so that it doesn't get run again when promoting local
      // // changes to that environment
      const initMigration = fs.readdirSync(migrationDirectory)[0];
      const version = initMigration.match(/^\d+/)[0];
      command = `docker run --rm -v $(pwd):/hasuracli nhost/hasura-cli-docker migrate apply --version "${version}" --skip-execution ${commonOptions}`;
      await exec(command, { cwd: nhostDir });

      // create metadata from remote
      command = `docker run --rm -v $(pwd):/hasuracli ${dockerImage}  metadata export ${commonOptions}`;
      await exec(command, { cwd: nhostDir });

      // create seeds from remote
      // auth.roles and auth.providers plus any enum compatible tables that might exist
      // all enum compatible tables must contain at least one row
      // https://hasura.io/docs/1.0/graphql/core/schema/enums.html#creating-an-enum-compatible-table
      let seedTables = ["auth.roles", "auth.providers"];

      // use the API to check whether this project has enum compatible tables
      command = `curl -d '{"type": "export_metadata", "args": {}}' -H 'X-Hasura-Admin-Secret: ${adminSecret}' ${hasuraEndpoint}/v1/query`;
      try {
        const response = await exec(command);
        const tables = JSON.parse(response.stdout).tables;
        // filter enum compatible tables
        const enumTables = tables.filter((table) => table.is_enum);

        enumTables.forEach(({ table }) =>
          seedTables.push(`${table.schema}.${table.name}`)
        );
      } catch (err) {}

      const fromTables = seedTables.reduce(
        (all, current) => `${all} --from-table ${current}`,
        ""
      );
      command = `docker run --rm -v $(pwd):/hasuracli ${dockerImage} seeds create roles_and_providers ${fromTables} ${commonOptions}`;
      await exec(command, { cwd: nhostDir });

      const extensions = await this._getExtensions(hasuraEndpoint, adminSecret);
      const extensionsWriteToFile = extensions
        .map((extension) => {
          return `CREATE EXTENSION IF NOT EXISTS ${extension};\n`;
        })
        .join("");
      extensionsWriteToFile.concat("\n\n");

      // write extensions to file in the beginning of the file
      const sqlPath = `${migrationDirectory}/${initMigration}/up.sql`;
      var data = fs.readFileSync(sqlPath); //read existing contents into data
      var fd = fs.openSync(sqlPath, "w+");
      var buffer = Buffer.from(extensionsWriteToFile);
      fs.writeSync(fd, buffer, 0, buffer.length, 0); //write new data
      fs.writeSync(fd, data, 0, data.length, buffer.length); //append old data
      fs.close(fd);

      // write dev environment variables to .env.development
      await writeFile(
        envFile,
        project.project_env_vars
          .map((envVar) => `${envVar.name}=${envVar.dev_value}`)
          .join("\n"),
        { flag: "a" }
      );

      await writeFile(
        envFile,
        `\nREGISTRATION_CUSTOM_FIELDS=${project.hbp_REGISTRATION_CUSTOM_FIELDS}\n`,
        { flag: "a" }
      );

      if (project.backend_user_fields) {
        await writeFile(
          envFile,
          `JWT_CUSTOM_FIELDS=${project.backend_user_fields}\n`,
          { flag: "a" }
        );
      }

      if (project.hbp_DEFAULT_ALLOWED_USER_ROLES) {
        await writeFile(
          envFile,
          `DEFAULT_ALLOWED_USER_ROLES=${project.hbp_DEFAULT_ALLOWED_USER_ROLES}\n`,
          { flag: "a" }
        );
      }
      if (project.hbp_allowed_user_roles) {
        await writeFile(
          envFile,
          `ALLOWED_USER_ROLES=${project.hbp_allowed_user_roles}\n`,
          { flag: "a" }
        );
      }
    } catch (error) {
      this.log(`${chalk.red("Error!")} ${error.message}`);
      // spinner.fail();
      stopSpinner();
      this.log(`${chalk.red("Error!")} ${error.message}`);
      this.exit(1);
    }

    spinner.succeed();
    stopSpinner();

    this.log(`${chalk.green("Nhost project successfully initialized")}`);
  }
}

InitCommand.description = `Initialize current working directory as a Nhost project
...
Initialize current working directory as a Nhost project
`;

module.exports = InitCommand;
