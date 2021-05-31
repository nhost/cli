const { Command } = require("@oclif/command");
const fs = require("fs");
const fetch = require("node-fetch");
const chalk = require("chalk");
const util = require("util");
const exec = util.promisify(require("child_process").exec);
const exists = util.promisify(fs.exists);
const writeFile = util.promisify(fs.writeFile);
const appendFile = util.promisify(fs.appendFile);
const mkdir = util.promisify(fs.mkdir);
const yaml = require("js-yaml");
const os = require("os");

const spinnerWith = require("../util/spinner");
const selectProject = require("../util/projects");
const {
  authFileExists,
  readAuthFile,
  getCustomApiEndpoint,
  getNhostConfig,
} = require("../util/config");
const { validateAuth } = require("../util/login");
const checkForHasura = require("../util/dependencies");

class InitCommand extends Command {
  projectOnHBPV2(project) {
    return project.backend_version.includes("v2");
  }

  async _getRoles(hasuraEndpoint, adminSecret) {
    const command = `curl -d '{"type": "run_sql", "args": {"sql": "SELECT * FROM auth.roles;"}}' -H 'X-Hasura-Admin-Secret: ${adminSecret}' ${hasuraEndpoint}/v1/query`;
    try {
      const response = await exec(command);
      const data = JSON.parse(response.stdout).result.splice(1); // remove head (first row)
      const roles = data.map((row) => row[0]);
      return roles;
    } catch (error) {
      console.error(error);
      console.error("Error getting auth roles");
    }
  }

  async _getProviders(hasuraEndpoint, adminSecret) {
    const command = `curl -d '{"type": "run_sql", "args": {"sql": "SELECT * FROM auth.providers;"}}' -H 'X-Hasura-Admin-Secret: ${adminSecret}' ${hasuraEndpoint}/v1/query`;
    try {
      const response = await exec(command);
      const data = JSON.parse(response.stdout).result.splice(1); // remove head (first row)
      const providers = data.map((row) => row[0]);
      return providers;
    } catch (error) {
      console.error(error);
      console.error("Error getting auth providers");
    }
  }

  async _getExtensions(hasuraEndpoint, adminSecret) {
    const command = `curl -d '{"type": "run_sql", "args": {"sql": "SELECT * FROM pg_extension;"}}' -H 'X-Hasura-Admin-Secret: ${adminSecret}' ${hasuraEndpoint}/v1/query`;
    try {
      const response = await exec(command);
      const data = JSON.parse(response.stdout).result.splice(1); // remove head (first row)
      const extensions = data.map((row) => row[1]);
      return extensions;
    } catch (error) {
      console.error(error);
      console.error("Error getting extensions");
    }
  }

  async _writeToFileSync(filePath, data, position = "end") {
    if (!["start", "end"].includes(position)) {
      throw 'Position must be on of: "start", "end"';
    }

    const currentData = fs.readFileSync(filePath);
    var fd = fs.openSync(filePath, "w+");
    var buffer = Buffer.from(data);
    if (position === "end") {
      // prepend old data
      fs.writeSync(fd, currentData, 0, currentData.length, 0);
      fs.writeSync(fd, buffer, 0, buffer.length, currentData.length);
    } else if (position === "start") {
      //append old data
      fs.writeSync(fd, buffer, 0, buffer.length, 0); //write new data
      fs.writeSync(fd, currentData, 0, currentData.length, buffer.length);
    }
    fs.close(fd);
  }

  //update values within .env file. if value is not found it is written 
  async _setEnvValue(filePath, key, value) {
    let newEntry = `${key}=${value}`;
    
    if(fs.existsSync(filePath)){
      
      let ENV_VARS = fs.readFileSync(filePath, "utf8").split(os.EOL);
      
      let targetIndex = ENV_VARS.indexOf(ENV_VARS.find((line) => 
      { 
          return line.indexOf('=') >= 0 && line.split('=')[0].trim() == key
      }));

      targetIndex >= 0 ?  ENV_VARS.splice(targetIndex, 1, newEntry) : ENV_VARS.push(newEntry);
      
      fs.writeFileSync(filePath, ENV_VARS.join(os.EOL));
    } else{

      fs.appendFileSync(filePath, newEntry); 
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

    // create root nhost folder
    await mkdir(nhostDir);
    // .nhost is used for nhost specific configuration
    await mkdir(dotNhost);
    await writeFile(
      `${dotNhost}/nhost.yaml`,
      `project_id: ${selectedProjectId}`
    );

    // generate Nhost config
    const nhostConfig = getNhostConfig(project);

    const safeDumpOptions = {
      skipInvalid: true,
    };
    const nhostConfigYaml = yaml.safeDump(nhostConfig, safeDumpOptions);

    // config.yaml holds configuration for GraphQL engine, PostgreSQL and HBP
    // it is also a requirement for hasura to work
    await writeFile(`${nhostDir}/config.yaml`, nhostConfigYaml);

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

    await writeFile(ignoreFile, "\n.nhost\napi/node_modules", {
      flag: "a",
    });

    // .env.development
    const envFile = `${workingDir}/.env.development`;
    if (await !exists(envFile)) {
      await writeFile(envFile, "# env vars from Nhost\n");
    }

    const hasuraEndpoint = `https://${project.project_domain.hasura_domain}`;
    const adminSecret = project.hasura_gqe_admin_secret;
    // const remoteHasuraVersion = project.hasura_gqe_version;
    // const dockerImage = `nhost/hasura-cli-docker:${remoteHasuraVersion}`;
    const hasuraCLI = `hasura`;
    const commonOptions = `--endpoint ${hasuraEndpoint} --admin-secret ${adminSecret} --skip-update-check`;

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

      // create migrations from remote
      spinner.text = "Create migrations";
      let command = `${hasuraCLI} migrate create "init" --from-server --schema "public" --schema "auth" ${commonOptions}`;
      await exec(command, { cwd: nhostDir });

      // // mark this migration as applied (--skip-execution) on the remote server
      // // so that it doesn't get run again when promoting local
      // // changes to that environment
      const initMigration = fs.readdirSync(migrationDirectory)[0];
      const version = initMigration.match(/^\d+/)[0];
      command = `${hasuraCLI} migrate apply --version "${version}" --skip-execution ${commonOptions}`;
      await exec(command, { cwd: nhostDir });

      // create metadata from remote
      spinner.text = "Create Hasura metadata";
      command = `${hasuraCLI} metadata export ${commonOptions}`;
      await exec(command, { cwd: nhostDir });

      // auth.roles and auth.providers plus any enum compatible tables that might exist
      // all enum compatible tables must contain at least one row
      // https://hasura.io/docs/1.0/graphql/core/schema/enums.html#creating-an-enum-compatible-table
      let seedTables = [];

      // use the API to check whether this project has enum compatible tables
      spinner.text = "Adding enum tables";
      command = `curl -d '{"type": "export_metadata", "args": {}}' -H 'X-Hasura-Admin-Secret: ${adminSecret}' ${hasuraEndpoint}/v1/query`;
      try {
        const response = await exec(command);
        const tables = JSON.parse(response.stdout).tables;
        // filter enum compatible tables
        const enumTables = tables.filter((table) => table.is_enum);
        enumTables.forEach(({ table }) => {
          seedTables.push(`${table.schema}.${table.name}`);
        });
      } catch (err) {}

      const fromTables = seedTables.reduce(
        (all, current) => `${all} --from-table ${current}`,
        ""
      );
      if (fromTables) {
        command = `${hasuraCLI} seeds create roles_and_providers ${fromTables} ${commonOptions}`;
        await exec(command, { cwd: nhostDir });
      }

      // add extensions to init migration
      spinner.text = "Add Postgres extensions to init migration";
      const extensions = await this._getExtensions(hasuraEndpoint, adminSecret);
      const extensionsWriteToFile = extensions
        .map((extension) => {
          return `CREATE EXTENSION IF NOT EXISTS ${extension};\n`;
        })
        .join("");
      extensionsWriteToFile.concat("\n\n");
      const sqlPath = `${migrationDirectory}/${initMigration}/up.sql`;
      this._writeToFileSync(sqlPath, extensionsWriteToFile, "start");

      // add auth.roles to init migration
      spinner.text = "Add auth roles to init migration";
      const roles = await this._getRoles(hasuraEndpoint, adminSecret);
      let rolesSQL = `\n\nINSERT INTO auth.roles (role)\n    VALUES `;
      const rolesMap = roles.map((role) => `('${role}')`).join(", ");
      rolesSQL += `${rolesMap};\n\n`;
      this._writeToFileSync(sqlPath, rolesSQL, "end");

      // add auth.providers to init migration
      spinner.text = "Add auth providers to init migration";
      const providers = await this._getProviders(hasuraEndpoint, adminSecret);
      let providersSQL = `INSERT INTO auth.providers (provider)\n    VALUES `;
      const providersMap = providers
        .map((provider) => `('${provider}')`)
        .join(", ");
      providersSQL += `${providersMap};\n\n`;
      this._writeToFileSync(sqlPath, providersSQL, "end");

      // write ENV variables to .env.development
      spinner.text = "Adding env vars to .env.development";
      await project.project_env_vars
           .reduce( async (memo, envVar) => { 
             await memo; 
             await this._setEnvValue(envFile, envVar.name, envVar.dev_value);
           }, undefined);

      await this._setEnvValue(
        envFile,
        'REGISTRATION_CUSTOM_FIELDS',
        project.hbp_REGISTRATION_CUSTOM_FIELDS
      );

      if (project.backend_user_fields) {
        await this._setEnvValue(
          envFile,
          'JWT_CUSTOM_FIELDS',
          project.backend_user_fields
        );
      }

      if (project.hbp_DEFAULT_ALLOWED_USER_ROLES) {
        await this._setEnvValue(
          envFile,
          'DEFAULT_ALLOWED_USER_ROLES',
          project.hbp_DEFAULT_ALLOWED_USER_ROLES
        );
      }
      if (project.hbp_allowed_user_roles) {
        await this._setEnvValue(
          envFile,
          'ALLOWED_USER_ROLES',
          project.hbp_allowed_user_roles
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
