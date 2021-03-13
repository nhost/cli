const { Command } = require("@oclif/command");
const fs = require("fs");
const { spawn, execSync } = require("child_process");
const yaml = require("js-yaml");
const crypto = require("crypto");
const chalk = require("chalk");
const nunjucks = require("nunjucks");
const kill = require("tree-kill");
const detect = require("detect-port");

const spinnerWith = require("../util/spinner");
const getComposeTemplate = require("../util/compose");
const getDockerApiTemplate = require("../util/docker-api");

const util = require("util");
const readFile = util.promisify(fs.readFile);
const exec = util.promisify(require("child_process").exec);
const exists = util.promisify(fs.exists);
const writeFile = util.promisify(fs.writeFile);
const unlink = util.promisify(fs.unlink);

let hasuraConsoleSpawn;
let startupFinished = false;

async function cleanup(path, errorMessage) {
  let { spinner } = spinnerWith("stopping Nhost");

  if (!startupFinished) {
    console.log(`\nWriting logs to ${path}/nhost.log\n`);
    await exec(
      `docker-compose -f ${path}/docker-compose.yaml logs --no-color -t  > ${path}/nhost.log`
    ).catch((error) =>
      console.log(
        `${chalk.red(`\nError during writing of logfile`)}\n\n${error}`
      )
    );
  }

  if (hasuraConsoleSpawn && hasuraConsoleSpawn.pid) {
    kill(hasuraConsoleSpawn.pid);
  }

  await exec(
    `docker-compose -f ${path}/docker-compose.yaml down`
  ).catch((error) =>
    console.log(
      `${chalk.red(`\nError during docker compose down`)}\n\n${error}`
    )
  );

  await unlink(`${path}/docker-compose.yaml`);
  await unlink(`${path}/Dockerfile-api`);
  if (startupFinished) {
    spinner.succeed("see you soon");
  } else {
    spinner.fail(errorMessage);
  }
  process.exit();
}

class DevCommand extends Command {
  async waitForGraphqlEngine(nhostConfig, timesRemaining = 300) {
    return new Promise((resolve, reject) => {
      const retry = (timesRemaining) => {
        try {
          execSync(
            `curl http://localhost:${nhostConfig.hasura_graphql_port}/healthz > /dev/null 2>&1`
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
    const workingDir = ".";
    const nhostDir = `${workingDir}/nhost`;
    const dotNhost = `${workingDir}/.nhost`;

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

    const firstRun = !(await exists(`${dotNhost}/db_data`));
    let startMessage = "Nhost is starting...";
    if (firstRun) {
      startMessage += `${chalk.bold.underline("first run takes longer")}`;
    }

    let { spinner, stopSpinner } = spinnerWith(startMessage);

    process.on("SIGINT", () => {
      stopSpinner();
      cleanup(dotNhost, "interrupted by signal");
    });

    const nhostConfig = yaml.safeLoad(
      await readFile(`${nhostDir}/config.yaml`, { encoding: "utf8" })
    );

    const ports = [
      "hasura_graphql_port",
      "hasura_backend_plus_port",
      "postgres_port",
      "minio_port",
      "api_port",
    ].map((p) => nhostConfig[p]);
    ports.push(9695);
    const freePorts = await Promise.all(ports.map((p) => detect(p)));
    const occupiedPorts = ports.filter((x) => !freePorts.includes(x));

    if (occupiedPorts.length > 0) {
      spinner.fail(
        `The following ports are not free, please change the nhost/config.yaml or stop the services: ${occupiedPorts}`
      );
      process.exit(1);
    }

    if (await exists("./api")) {
      nhostConfig["startApi"] = true;
    }

    nhostConfig.graphql_jwt_key = crypto
      .randomBytes(128)
      .toString("hex")
      .slice(0, 128);

    await writeFile(
      `${dotNhost}/docker-compose.yaml`,
      nunjucks.renderString(getComposeTemplate(), nhostConfig)
    );

    // write docker api file
    await writeFile(`${dotNhost}/Dockerfile-api`, getDockerApiTemplate());

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
      this.log(`${chalk.red("Error!")} ${err.message}`);
      stopSpinner();
      cleanup(dotNhost, "Failed to start docker-compose");
    }

    // check whether GraphQL engine is up & running
    try {
      await this.waitForGraphqlEngine(nhostConfig);
    } catch (err) {
      spinner.fail();
      this.log(`${chalk.red("Nhost could not start!")} ${err.message}`);
      stopSpinner();
      cleanup(dotNhost, "Failed to start GraphQL Engine");
    }

    if (firstRun && fs.readdirSync(`${nhostDir}/seeds`).length > 0) {
      try {
        spinner.text = "Applying seed data";
        await exec(
          `hasura seeds apply --admin-secret ${nhostConfig.hasura_graphql_admin_secret}`,
          { cwd: nhostDir }
        );
      } catch (err) {
        spinner.fail();
        this.log(`${chalk.red("Error!")} ${err.message}`);
        stopSpinner();
        cleanup(dotNhost, "Failed to start apply seeds");
      }
    }

    try {
      spinner.text = "Applying metadata";
      await exec(
        `hasura metadata apply --admin-secret ${nhostConfig.hasura_graphql_admin_secret}`,
        { cwd: nhostDir }
      );
    } catch (err) {
      spinner.fail();
      this.log(`${chalk.red("Error!")} ${err.message}`);
      stopSpinner();
      cleanup(dotNhost, "Failed to start apply metadata");
    }

    hasuraConsoleSpawn = spawn(
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
      `Local Nhost backend is up!
GraphQL API:\t${chalk.underline.bold(
        `http://localhost:${nhostConfig.hasura_graphql_port}/v1/graphql`
      )}
Hasura Console:\t${chalk.underline.bold("http://localhost:9695")}
Auth & Storage:\t${chalk.underline.bold(
        `http://localhost:${nhostConfig.hasura_backend_plus_port}`
      )}
Custom API:\t\t${chalk.underline.bold(
        `http://localhost:${nhostConfig.api_port}`
      )}`
    );

    stopSpinner();
    startupFinished = true;
  }
}

DevCommand.description = `Start Nhost project for local development
...
Start Nhost project for local development
`;

nunjucks.configure({ autoescape: true });

module.exports = DevCommand;
