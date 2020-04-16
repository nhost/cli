const { Command, flags } = require("@oclif/command");
const fs = require("fs");
const { execSync } = require("child_process");
const moveTemplateMigration = require("../migrations");
const yaml = require("js-yaml");

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

# custom environment variables for Hasura GraphQL engine: webhooks, etc
env_file: .env.development
`;
    return configData;
  }

  async run() {
    const { flags } = this.parse(InitCommand);
    let directory = flags.directory;
    const endpoint = flags.endpoint;
    const adminSecret = flags["admin-secret"];

    // check for Hasura CLI
    try {
      execSync("command -v hasura");
    } catch {
      return this.warn(
        "Hasura CLI is a dependency. Please follow the instructions here https://hasura.io/docs/1.0/graphql/manual/hasura-cli/install-hasura-cli.html"
      );
    }

    if (adminSecret && !endpoint) {
      return this.warn(
        "When using --admin-secret, --endpoint also needs to be specified"
      );
    }

    if (endpoint && directory) {
      return this.warn(
        "When initialising from an existing project on Nhost, please run the command within your target directory, without using -d"
      );
    }

    if (directory) {
      if (!fs.existsSync(directory)) {
        fs.mkdirSync(directory);
      } else {
        return this.warn(
          "For an existing directory, please run 'nhost init' within it"
        );
      }
    } else {
      // assume current working directory if no directory is provided through -d
      directory = ".";
    }

    // config.yaml has various configuration for GraphQL engine, PostgreSQL and HBP
    // it is also a requirement for the Hasura CLI to run commands - can't be renamed
    fs.writeFileSync(`${directory}/config.yaml`, this.getConfigData());

    // create a migrations directory if not present
    const migrationDirectory = `${directory}/migrations`;
    if (!fs.existsSync(migrationDirectory)) {
      fs.mkdirSync(migrationDirectory);
    }

    // create or append to .gitignore
    const ignoreFile = `${directory}/.gitignore`;
    fs.writeFileSync(ignoreFile, "\nconfig.yaml\n.nhost\ndb_data", {
      flag: "a",
    });

    // if --endpoint is provided it means an existing project is being used
    if (endpoint) {
      let command = `hasura migrate create "init" --from-server --endpoint ${endpoint} --schema "public" --schema "auth"`;
      if (adminSecret) {
        command += ` --admin-secret ${adminSecret}`;
      }

      try {
        execSync(command, { stdio: "inherit" });

        const initMigration = fs.readdirSync("./migrations")[0];
        const metadata = yaml.safeLoad(
          fs.readFileSync(`./migrations/${initMigration}/up.yaml`, {
            encoding: "utf8",
          })
        );

        // TODO: rethink this implementation
        // fragile because it relies on Hasura metadata format
        // there are 2 places where ENV vars might be used with event triggers
        const eventTriggers = metadata[0].args.tables
          .filter((table) => table.event_triggers)
          .flatMap((table) => table.event_triggers);

        // (1) webhook URL (webhook_from_env)
        const webhooksFromEnv = eventTriggers
          .filter((trigger) => trigger.webhook_from_env)
          .map((trigger) => `${trigger.webhook_from_env}=changeme`)
          .filter((value, index, self) => {
            // remove duplicates if any (same webhook env var for multiple events)
            return self.indexOf(value) === index;
          });

        // (2) headers (value_from_env)
        const headersFromEnv = eventTriggers
          .filter((trigger) => trigger.headers)
          .flatMap((trigger) => trigger.headers)
          .filter((header) => header.value_from_env)
          .map((header) => `${header.value_from_env}=changeme`)
          .filter((value, index, self) => {
            // remove duplicates if any (same header env var for multiple events)
            return self.indexOf(value) === index;
          });

        if (webhooksFromEnv.length > 0 || headersFromEnv.length > 0) {
          fs.writeFileSync(
            `${directory}/.env.development`,
            webhooksFromEnv.concat(headersFromEnv).join("\n"),
            { flag: "a" }
          );

          fs.writeFileSync(ignoreFile, "\n.env.development", { flag: "a" });
        }
      } catch (error) {
        this.error("Something went wrong: ", error);
      }
    } else {
      // when no endpoint is specified, we ship a template
      // migration based on HBP most up-to-date schema
      moveTemplateMigration(migrationDirectory);
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
Initialises a new project (or an existing one) with configuration for running the Nhost environment
`;

InitCommand.flags = {
  directory: flags.string({
    char: "d",
    description: "Where to create your project",
    required: false,
  }),
  endpoint: flags.string({
    char: "e",
    description: "Endpoint of your GraphQL engine running on Nhost",
    required: false,
  }),
  "admin-secret": flags.string({
    char: "a",
    description: "GraphQL engine admin secret",
    required: false,
  }),
};

module.exports = InitCommand;
