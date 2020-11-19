const { Command } = require("@oclif/command");
const fs = require("fs");
const { spawn, execSync } = require("child_process");
const yaml = require("js-yaml");
const crypto = require("crypto");
const chalk = require("chalk");
const nunjucks = require("nunjucks");

const spinnerWith = require("../util/spinner");
const getComposeTemplate = require("../util/compose");
const getDockerApiTempalte = require("../util/docker-api");

const util = require("util");
const readFile = util.promisify(fs.readFile);
const exec = util.promisify(require("child_process").exec);
const exists = util.promisify(fs.exists);
const writeFile = util.promisify(fs.writeFile);
const unlink = util.promisify(fs.unlink);

async function cleanup(path = "./nhost/.nhost") {
  let { spinner } = spinnerWith("stopping Nhost");

  await exec(`docker-compose -f ${path}/docker-compose.yaml down`);
  await unlink(`${path}/docker-compose.yaml`);
  spinner.succeed("see you soon");
  process.exit();
}

class DevCommand extends Command {
  async waitForGraphqlEngine(nhostConfig, timesRemaining = 60) {
    return new Promise((resolve, reject) => {
      const retry = (timesRemaining) => {
        try {
          execSync(
            `curl http://localhost:${nhostConfig.hasura_graphql_port}/v1/version > /dev/null 2>&1`
          );

          return resolve();
        } catch (err) {
          if (timesRemaining === 0) {
            return reject(err);
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
    process.on("SIGINT", () => cleanup());
    const workingDir = ".";
    const nhostDir = `${workingDir}/nhost`;

    if (!(await exists(nhostDir))) {
      return this.log(
        `${chalk.red(
          "Error!"
        )} initialize your project before with ${chalk.bold.underline(
          "nhost init"
        )} or make sure to run commands at the root of your project`
      );
    }

    // check if docker-compose is installed
    try {
      await exec("command -v docker-compose");
    } catch {
      return this.log(
        `${chalk.red("Error!")} please make sure to have ${chalk.bold.underline(
          "docker compose"
        )} installed`
      );
    }

    const dbIncluded = !(await exists(`${nhostDir}/db_data`));
    let startMessage = "Nhost is starting...";
    if (dbIncluded) {
      startMessage += `${chalk.bold.underline("first run takes longer")}`;
    }

    let { spinner, stopSpinner } = spinnerWith(startMessage);

    const nhostConfig = yaml.safeLoad(
      await readFile(`${nhostDir}/config.yaml`, { encoding: "utf8" })
    );

    nhostConfig.graphql_jwt_key = crypto
      .randomBytes(128)
      .toString("hex")
      .slice(0, 128);

    // create .nhost
    const dotNhost = `${nhostDir}/.nhost`;
    await writeFile(
      `${dotNhost}/docker-compose.yaml`,
      nunjucks.renderString(getComposeTemplate(), nhostConfig)
    );

    // write docker api file
    await writeFile(`${dotNhost}/Dockerfile-api`, getDockerApiTempalte());

    // validate compose file
    await exec(`docker-compose -f ${dotNhost}/docker-compose.yaml config`);

    // run docker-compose up
    try {
      await exec(
        `docker-compose -f ${dotNhost}/docker-compose.yaml up -d --build`
      );
    } catch (err) {
      spinner.fail();
      this.log(`${chalk.red("Error!")} ${err.message}`);
      stopSpinner();
      cleanup();
    }

    // check whether GraphQL engine is up & running
    try {
      await this.waitForGraphqlEngine(nhostConfig);
    } catch (err) {
      spinner.fail();
      this.log(`${chalk.red("Nhost could not start!")} ${err.message}`);
      stopSpinner();
      cleanup();
    }

    if (dbIncluded) {
      try {
        await exec(
          `hasura seeds apply --admin-secret ${nhostConfig.hasura_graphql_admin_secret}`,
          { cwd: nhostDir }
        );
      } catch (err) {
        spinner.fail();
        this.log(`${chalk.red("Error!")} ${err.message}`);
        stopSpinner();
        cleanup();
        this.exit();
      }
    }

    spawn(
      "hasura",
      [
        "console",
        `--endpoint=http://localhost:${nhostConfig.hasura_graphql_port}`,
        `--admin-secret=${nhostConfig.hasura_graphql_admin_secret}`,
        "--console-port=9695",
      ],
      { stdio: "ignore", cwd: nhostDir }
    );

    spinner.succeed(
      `Nhost is running! The hasura console can be found at ${chalk.underline.bold(
        "http://localhost:9695"
      )}`
    );

    stopSpinner();
  }
}

DevCommand.description = `Start Nhost project for local development
...
Start Nhost project for local development
`;

nunjucks.configure({ autoescape: true });

module.exports = DevCommand;
