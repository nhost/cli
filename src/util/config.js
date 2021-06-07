const fs = require("fs");
const chalk = require("chalk");
const util = require("util");
const exists = util.promisify(fs.exists);
const unlink = util.promisify(fs.unlink);
const path = require("path");
const { homedir } = require("os");
const writeJSON = require("write-json-file");

const NHOST_DIR = path.join(homedir(), ".nhost");
const authPath = path.join(NHOST_DIR, "auth.json");

let configData = `# hasura configuration for the CLI
version: 2
metadata_directory: metadata

# hasura configuration for the project
hasura_graphql_version: {{ hasura_gqe_version }}
hasura_graphql_port: 8080
hasura_graphql_admin_secret: 123456

# hasura backend plus configuration
hasura_backend_plus_version: {{ backend_version }}
hasura_backend_plus_port: 9001

# postgres configuration
postgres_version: {{ postgres_version }}
postgres_port: 5432
postgres_user: postgres
postgres_password: postgres
postgres_image: postgres

# minio
minio_port: 9000

# api
api_port: 4000

# custom environment variables for Hasura GraphQL engine: webhooks, headers, etc
env_file: ../.env.development

# OAuth services
provider_success_redirect: http://localhost:3000
provider_failure_redirect: http://localhost:3000/login-fail

google_enable: false
google_client_id:
google_client_secret:

github_enable: false
github_client_id:
github_client_secret:

facebook_enable: false
facebook_client_id:
facebook_client_secret:

linkedin_enable: false
linkedin_client_id:
linkedin_client_secret:
`;

async function writeAuthFile(data) {
  try {
    return writeJSON.sync(authPath, data, {
      indent: 2,
      mode: 0o600,
    });
  } catch (err) {
    console.log(chalk.bold.red("Error!"));
    console.log("Could not read auth file. Run `nhost login` first.");
    throw err;
  }
}

async function deleteAuthFile() {
  try {
    await unlink(authPath);
  } catch (err) {
    throw err;
  }
}

function readAuthFile() {
  try {
    return require(authPath);
  } catch (error) {
    console.log(chalk.bold.red("Error!"));
    console.log("Could not read auth file. Run `nhost login` first.");
    throw error;
  }
}

async function authFileExists() {
  return await exists(authPath);
}

function getCustomApiEndpoint() {
  return "https://customapi.nhost.io";
}

function getNhostConfig(options) {
  const { hasura_gqe_version, backend_version, postgres_version } = options;

  const nhostConfig = {
    version: 2,
    metadata_directory: "metadata",
    hasura_graphql_version: hasura_gqe_version,
    hasura_graphql_port: 8080,
    hasura_graphql_admin_secret: 123456,
    hasura_backend_plus_version: backend_version,
    hasura_backend_plus_port: 9001,
    postgres_version: postgres_version,
    postgres_port: 5432,
    postgres_user: "postgres",
    postgres_password: "postgres",
    minio_port: 9000,
    api_port: 4000,
    env_file: "../.env.development",
    provider_success_redirect: "http://localhost:3000",
    provider_failure_redirect: "http://localhost:3000/login-fail",
    google_enable: false,
    google_client_id: "",
    google_client_secret: "",
    github_enable: false,
    github_client_id: "",
    github_client_secret: "",
    facebook_enable: false,
    facebook_client_id: "",
    facebook_client_secret: "",
    linkedin_enable: false,
    linkedin_client_id: "",
    linkedin_client_secret: "",
  };

  return nhostConfig;
}

module.exports = {
  writeAuthFile,
  readAuthFile,
  authFileExists,
  getCustomApiEndpoint,
  getNhostConfig,
  deleteAuthFile,
};
