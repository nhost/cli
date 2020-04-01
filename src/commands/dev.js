const { Command, flags } = require("@oclif/command");
const fs = require("fs");
const { execSync } = require("child_process");
const yaml = require("js-yaml");
const nunjucks = require("nunjucks");
const crypto = require("crypto");

class DevCommand extends Command {
  async run() {
    if (
      !fs.existsSync("./docker-compose.example") ||
      !fs.existsSync("./config.yaml")
    ) {
      return this.log(
        "Please run `nhost init` before starting a development environment."
      );
    }

    const nhostConfig = yaml.safeLoad(
      fs.readFileSync("config.yaml", { encoding: "utf8" })
    );
    const dockerComposeTemplate = fs.readFileSync("docker-compose.example", {
      encoding: "utf8"
    });

    const jwtSecret = crypto
      .randomBytes(128)
      .toString("hex")
      .slice(0, 128);

    nhostConfig.graphql_jwt_key = jwtSecret;
    fs.writeFileSync(
      "docker-compose.yaml",
      nunjucks.renderString(dockerComposeTemplate, nhostConfig)
    );

    const dockerFirstRun = !fs.existsSync("./db_data");

    // validate compose file
    execSync("docker-compose -f ./docker-compose.yaml config");
    execSync("docker-compose up -d > /dev/null 2>&1");

    this.log(
      `development environment is launching...console will be launched at http://localhost:${nhostConfig.graphql_server_port}`
    );

    if (dockerFirstRun) {
      this.log(
        "This seems to be the first time running nhost dev in this project so it might take longer to start..."
      );
    }

    // check whether the graphql-engine endpoint is up & running
    let reachable = false;
    while (!reachable) {
      try {
        execSync(
          `hasura console --endpoint=http://localhost:${nhostConfig.graphql_server_port} --admin-secret=${nhostConfig.graphql_admin_secret} > /dev/null 2>&1`
        );
      } catch (Error) {
        continue;
      }
    }
  }
}

DevCommand.description = `Describe the command here
...
Extra documentation goes here
`;

DevCommand.flags = {
  name: flags.string({ char: "n", description: "name to print" })
};

nunjucks.configure({ autoescape: true });

module.exports = DevCommand;
