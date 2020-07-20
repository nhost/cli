const { Command, flags } = require("@oclif/command");
const fs = require("fs");
const { spawn, execSync } = require("child_process");
const yaml = require("js-yaml");
const crypto = require("crypto");
const chalk = require("chalk");
const nunjucks = require("nunjucks");

const spinnerWith = require("../util/spinner");
const getComposeTemplate = require("../util/compose");

const util = require("util");
const readFile = util.promisify(fs.readFile);
const exec = util.promisify(require("child_process").exec);
const exists = util.promisify(fs.exists);
const mkdir = util.promisify(fs.mkdir);
// const rmdir = util.promisify(fs.rmdir);
const writeFile = util.promisify(fs.writeFile);

function cleanup(path = "./.nhost") {
  console.log(chalk.white("Nhost is shutting down"));
  execSync(
    `docker-compose -f ${path}/docker-compose.yaml down > /dev/null 2>&1`
  );
  fs.rmdirSync(path, { recursive: true });
  process.exit();
}

class DevCommand extends Command {
  async waitForGraphqlEngine(nhostConfig, timesRemaining = 60) {
    return new Promise((resolve, reject) => {
      const retry = (timesRemaining) => {
        try {
          execSync(
            `curl -X GET http://localhost:${nhostConfig.graphql_server_port}/v1/version > /dev/null 2>&1`
          );

          return resolve();
        } catch (error) {
          if (timesRemaining === 0) {
            return reject();
          }

          setTimeout(() => {
            retry(--timesRemaining);
          }, 1000);
        }
      };

      retry(timesRemaining);
    });
  }

  async run() {
    if (!await exists("./config.yaml")) {
      return this.log(
        `${chalk.red(
          "Error!"
        )} initialize your project before running ${chalk.bold.underline(
          "nhost dev"
        )}`
      );
    }

    // check if docker-compose is installed
    try {
      await exec("command -v docker-compose");
    } catch {
      this.log(
        `${chalk.red("Error!")} please make sure to have ${chalk.bold.underline(
          "docker compose"
        )} installed`
      );
    }

    const firstRun = !await exists("./db_data");
    let startMessage = "Nhost is starting";
    if (firstRun) {
      startMessage += `, ${chalk.bold.underline("database included")}`;
    }
    
    let { spinner, stopSpinner } = spinnerWith(startMessage);

    const nhostConfig = yaml.safeLoad(
      await readFile("./config.yaml", { encoding: "utf8" })
    );

    // generate random admin secret if not specified in config.yaml
    if (!nhostConfig.graphql_admin_secret) {
      nhostConfig.graphql_admin_secret = crypto
        .randomBytes(32)
        .toString("hex")
        .slice(0, 32);
    }
    nhostConfig.graphql_jwt_key = crypto
      .randomBytes(128)
      .toString("hex")
      .slice(0, 128);

    // create .nhost
    const tempDir = "./.nhost";
    await mkdir(tempDir);

    await writeFile(
      `${tempDir}/docker-compose.yaml`,
      nunjucks.renderString(getComposeTemplate(), nhostConfig)
    );

    // validate compose file
    await exec(`docker-compose -f ${tempDir}/docker-compose.yaml config`);

    // try running docker-compose up
    try {
      await exec(
        // `docker-compose -f ${tempDir}/docker-compose.yaml up -d > /dev/null 2>&1`
        `docker-compose -f ${tempDir}/docker-compose.yaml up -d`
      );
    } catch {
      // TODO: improve error handling/messaging
      // issues here, after validation, are about ports not being available
      this.error("Please make sure all ports in 'config.yaml' are available");
    }

    // check whether GraphQL engine is up & running
    await this.waitForGraphqlEngine(nhostConfig)
      .then(() => {
        // launch hasura console and inherit stdio/stdout/stderr
        spawn(
          "hasura",
          [
            "console",
            `--endpoint=http://localhost:${nhostConfig.graphql_server_port}`,
            `--admin-secret=${nhostConfig.graphql_admin_secret}`,
          ],
          { stdio: "pipe" }
        );
      })
      .catch(() => {
        this.log(
          "Nhost could not start. Please make sure that all configuration is correct"
        );
        stopSpinner();
        cleanup();
      });

    spinner.succeed("Nhost is running");
    this.log(
      `Hasura console is running at ${chalk.underline.bold(
        "http://localhost:9695"
      )}`
    );
    stopSpinner();
  }
}

DevCommand.description = `Starts Nhost local development
...
Starts a complete Nhost environment with PostgreSQL, Hasura GraphQL Engine and Hasura Backend Plus (HBP)
`;

DevCommand.flags = {
  name: flags.string({ char: "n", description: "name to print" }),
};

process.on("SIGINT", () => {
  cleanup();
});

nunjucks.configure({ autoescape: true });

module.exports = DevCommand;
