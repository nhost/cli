const { Command, flags } = require("@oclif/command");
const fs = require("fs");
const { execSync } = require("child_process");
const moveTemplateMigration = require("../migrations");

class InitCommand extends Command {
  getConfigData() {
    let configData = `# configurations used when starting your environment

# hasura graphql configuration
graphql_version: 'v1.1.0.cli-migrations'
graphql_server_port: 8080
#graphql_admin_secret: (optional: if not specified, it will be generated on 'nhost dev')

# postgres configuration
postgres_version: 12.0
postgres_port: 5432
postgres_user: postgres
postgres_password: postgres
#postgres_db_data: (optional: if not specified, './db_data' will be used) 

# hasura backend plus configuration
backend_plus_version: v1.2.3
backend_plus_port: 9000
`;
    return configData;
  }

  async run() {
    const { flags } = this.parse(InitCommand);
    let directory = flags.directory;
    const endpoint = flags.endpoint;
    const adminSecret = flags["admin-secret"];

    // check if hasura's CLI is installed
    try {
      execSync("command -v hasura");
    } catch {
      this.error(
        "Hasura CLI is a dependency. Please follow the instructions here https://hasura.io/docs/1.0/graphql/manual/hasura-cli/install-hasura-cli.html"
      );
    }

    if (adminSecret && !endpoint) {
      return this.log("Please specify an endpoint with --endpoint");
    }

    if (directory) {
      if (!fs.existsSync(directory)) {
        fs.mkdirSync(directory);
      } else {
        return this.log(
          "For existing directories please run `nhost init` inside"
        );
      }
    } else {
      // if no directory is provided through the -d option, assume current working directory
      directory = ".";
    }

    const nhostConfigFile = `${directory}/config.yaml`;
    fs.writeFileSync(nhostConfigFile, this.getConfigData());

    // create the migrations directory if not present
    const migrationDirectory = `${directory}/migrations`;
    if (!fs.existsSync(migrationDirectory)) {
      fs.mkdirSync(migrationDirectory);
    }

    // if --endpoint is provided it means an existing project is being used
    if (endpoint) {
      let command = `hasura migrate create "init" --from-server --endpoint ${endpoint}`;
      if (adminSecret) {
        command += ` --admin-secret ${adminSecret}`;
      }

      try {
        execSync(command, { stdio: "inherit" });
      } catch (error) {
        this.error("Something went wrong: ", error);
      }

      const version = fs.readdirSync("./migrations")[0].match(/^[0-9]+/)[0];
      command = `hasura migrate apply --version "${version}" --skip-execution`;
      if (adminSecret) {
        command += ` --admin-secret ${adminSecret}`;
      }

      execSync(command, { stdio: "inherit" });
    } else {
      moveTemplateMigration(migrationDirectory);
    }

    const ignoreFile = `${directory}/.gitignore`;
    if (fs.existsSync(ignoreFile)) {
      execSync(`echo "\nconfig.yaml\n.nhost\ndb_data" >> ${ignoreFile}`);
    } else {
      execSync(`echo config.yaml > ${ignoreFile}`);
    }

    let initMessage = "Nhost boilerplate created";
    if (directory != ".") {
      initMessage += ` within ${directory}`;
    }

    this.log(initMessage);
  }
}

InitCommand.description = `Prepares a project to run with Nhost
...
Initializes a new project (or an existing one) with configuration for running the Nhost environment
`;

InitCommand.flags = {
  directory: flags.string({
    char: "d",
    description: "Where to create your project",
    required: false,
  }),
  endpoint: flags.string({
    char: "e",
    description: "Endpoint where the current project is running",
    required: false,
  }),
  "admin-secret": flags.string({
    char: "a",
    description: "Admin Secret",
    required: false,
  }),
};

module.exports = InitCommand;
