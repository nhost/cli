const { Command, flags } = require("@oclif/command");
const fs = require("fs");
const { exec, execSync } = require("child_process");
const createAndPopulateMigrations = require("../migrations");

class InitCommand extends Command {
  getConfigData() {
    let configData = `# values here are used by 'nhost dev' to start your dev environment

# hasura graphql configuration
graphql_version: v1.1.0.cli-migrations
graphql_server_port: 8080
#graphql_admin_secret: 123456 (optional: if not specified, it will be handled on 'nhost dev')

# postgres configuration
postgres_version: 12.0
postgres_port: 5432
postgres_user: postgres
postgres_password: postgres
`;
    return configData;
  }

  async run() {
    const { flags } = this.parse(InitCommand);
    let directory = flags.directory;

    if (directory) {
      if (!fs.existsSync(directory)) {
        fs.mkdirSync(directory);
      } else {
        return this.log(
          "Directory already exists. Please run `nhost init` within it and without the -d option if intended."
        );
      }
    } else {
      // if no directory is provided through the -d option, assume current working directory
      directory = ".";
    }

    // create the migrations directory and copy the initial migration there
    createAndPopulateMigrations(`${directory}/migrations`);

    const nhostConfigFile = `${directory}/config.yaml`;
    fs.writeFileSync(nhostConfigFile, this.getConfigData());

    const ignoreFile = `${directory}/.gitignore`;
    if (fs.existsSync(ignoreFile)) {
      execSync(`echo config.yaml >> ${ignoreFile}`);
    } else {
      execSync(`echo config.yaml > ${ignoreFile}`);
    }

    // finally check if hasura's CLI is installed
    exec("command -v hasura", (error) => {
      if (error) {
        this.log(
          "The hasura CLI is a dependency. Please check out the installation instructions here https://hasura.io/docs/1.0/graphql/manual/hasura-cli/install-hasura-cli.html"
        );
      }
    });

    if (directory === ".") {
      this.log("Nhost boilerplace created!");
    } else {
      this.log(`Nhost boilerplate created within ${directory}!`);
    }
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
  }),
};

module.exports = InitCommand;
