const { Command } = require("@oclif/command");
const chalk = require("chalk");
const yaml = require("js-yaml");
const util = require("util");
const fs = require("fs");

const { validateAuth } = require("../../util/login");

class DownCommand extends Command {
  async run() {
    this.log(`Overwriting existing .env.development file`);
    const projectConfig = yaml.safeLoad(
      fs.readFileSync(`.nhost/nhost.yaml`, { encoding: "utf8" })
    );
    const projectId = projectConfig.project_id;

    // get env vars from remote
    const userData = await validateAuth();

    const projects = [
      ...userData.user.projects,
      ...userData.user.teams.flatMap(({ team }) => team.projects),
    ];

    const project = projects.find((project) => project.id === projectId);
    this.log(
      `Downloading development environment variables for project: ${project.name}`
    );

    const envFile = `.env.development`;

    var envFileContent = fs.readFileSync(envFile, { encoding: "utf8" });

    const existingEnvVars = envFileContent
      .split("\n")
      .filter((row) => {
        return row;
      })
      .map((row) => {
        const [name, value] = row.split("=");
        return {
          name,
          value,
        };
      });

    const updatedProjectEnvVarIndexs = [];

    // update env vars already in .env.development
    const envVars = existingEnvVars.map((existingEnvVar) => {
      const i = project.project_env_vars.findIndex(
        (pEnvVar) => pEnvVar.name == existingEnvVar.name
      );
      if (i === -1) {
        return existingEnvVar;
      }

      const tmpEnvVar = project.project_env_vars[i];
      updatedProjectEnvVarIndexs.push(i);

      const res = {
        name: existingEnvVar.name,
        value: tmpEnvVar.dev_value,
      };
      return res;
    });

    // add env vars not already in .env.development
    project.project_env_vars
      .filter((_, i) => {
        // filter if the env var was alrady updated
        return !updatedProjectEnvVarIndexs.includes(i);
      })
      .forEach((envVar) => {
        // add new env var
        envVars.push({
          name: envVar.name,
          value: envVar.dev_value,
        });
      });

    fs.writeFileSync(
      envFile,
      envVars.map((envVar) => `${envVar.name}=${envVar.value}`).join("\n"),
      { flag: "w" }
    );

    this.log(`${chalk.white("✅  .env.development file updated")}`);
  }
}

DownCommand.description = `bbb :D`;

module.exports = DownCommand;
