const { Command, flags } = require("@oclif/command");
const fs = require("fs");
const { spawn, execSync } = require("child_process");
const yaml = require("js-yaml");
const nunjucks = require("nunjucks");
const crypto = require("crypto");

const dockerComposeTemplate = `version: '3.6'
services:
  nhost-postgres:
    image: postgres:{{ postgres_version }}
    ports:
      - '{{ postgres_port }}:{{ postgres_port }}'
    restart: always
    environment:
      POSTGRES_USER: {{ postgres_user }}
      POSTGRES_PASSWORD: {{ postgres_password }}
    volumes:
      - ../db_data:/var/lib/postgresql/data
  nhost-graphql-engine:
    image: hasura/graphql-engine:v1.1.0.cli-migrations
    ports:
      - '{{ graphql_server_port }}:{{ graphql_server_port }}'
    depends_on:
      - nhost-postgres
    restart: always
    environment:
      HASURA_GRAPHQL_SERVER_PORT: {{ graphql_server_port }}
      HASURA_GRAPHQL_DATABASE_URL: postgres://{{ postgres_user }}:{{ postgres_password }}@nhost-postgres:{{ postgres_port }}/postgres
      HASURA_GRAPHQL_ENABLE_CONSOLE: 'false'
      HASURA_GRAPHQL_ENABLED_LOG_TYPES: startup, http-log, webhook-log, websocket-log, query-log
      HASURA_GRAPHQL_ADMIN_SECRET: {{ graphql_admin_secret }}
      HASURA_GRAPHQL_JWT_SECRET: '{"type":"HS256", "key": "{{ graphql_jwt_key }}"}'
      HASURA_GRAPHQL_MIGRATIONS_SERVER_TIMEOUT: 5
      HASURA_GRAPHQL_NO_OF_RETRIES: 5
    command:
      - graphql-engine
      - serve
    volumes:
      - ../migrations:/hasura-migrations
  hasura-backend-plus:
    image: elitan/hasura-backend-plus:v1.1.1
    depends_on:
    - nhost-graphql-engine
    restart: always
    environment:
      PORT: 9000
      USER_FIELDS: ''
      USER_REGISTRATION_AUTO_ACTIVE: 'true'
      HASURA_GRAPHQL_ENDPOINT: http://nhost-graphql-engine:{{ graphql_server_port }}/v1/graphql
      HASURA_GRAPHQL_ADMIN_SECRET: {{ graphql_admin_secret }}
      HASURA_GRAPHQL_JWT_SECRET: '{"type":"HS256", "key": "{{ graphql_jwt_key }}"}'
      AUTH_ACTIVE: 'true'
      AUTH_LOCAL_ACTIVE: 'true'
      REFRESH_TOKEN_EXPIRES: 43200
      JWT_TOKEN_EXPIRES: 15
    ports:
      - 9000:9000
`;

class DevCommand extends Command {
  async run() {
    if (
      !fs.existsSync("./config.yaml")
    ) {
      return this.log(
        "Please run `nhost init` before starting a development environment."
      );
    }

    const nhostConfig = yaml.safeLoad(
      fs.readFileSync("./config.yaml", { encoding: "utf8" })
    );

    const graphqlAdminSecret = crypto
      .randomBytes(32)
      .toString("hex")
      .slice(0, 32);

    const jwtSecret = crypto
      .randomBytes(128)
      .toString("hex")
      .slice(0, 128);

    
    // create .nhost to hold the docker-compose file
    const tempDir = "./.nhost";
    if (!fs.existsSync(tempDir)) {
      fs.mkdirSync(tempDir);
    }

    nhostConfig.graphql_jwt_key = jwtSecret;
    nhostConfig.graphql_admin_secret = graphqlAdminSecret;
    fs.writeFileSync(
      `${tempDir}/docker-compose.yaml`,
      nunjucks.renderString(dockerComposeTemplate, nhostConfig)
    );

    const firstRun = !fs.existsSync("./db_data");
    // validate compose file
    execSync(`docker-compose -f ${tempDir}/docker-compose.yaml config`);
    execSync(`docker-compose -f ${tempDir}/docker-compose.yaml up -d > /dev/null 2>&1`);

    this.log(`development environment is launching...`);
    
    // additional warning because Postgres takes needs more time on its first startup (db_data)
    if (firstRun) {
      this.log(
        "This seems to be the first time running nhost dev in this project so it might take longer to start..."
      );
    }

    // check whether the graphql-engine is up & running
    let engineReachable = false;
    while (!engineReachable) {
      try {
        execSync(
          `curl -X GET http://localhost:${nhostConfig.graphql_server_port}/v1/version > /dev/null 2>&1`
        );
      } catch (Error) {
        continue;
      }

      engineReachable = true;
    }

    this.log(
      `ready...console is running at http://localhost:${nhostConfig.graphql_server_port}`
    );

    const consoleCommand = spawn(
      "hasura",
      [
        "console",
        `--endpoint=http://localhost:${nhostConfig.graphql_server_port}`,
        `--admin-secret=${nhostConfig.graphql_admin_secret}`
      ],
      { detached: true, stdio: "ignore" }
    );

    fs.writeFileSync("./.console.pid", consoleCommand.pid);

    consoleCommand.unref();
    this.log("to tear down your environment simply issue 'nhost destroy'");
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
