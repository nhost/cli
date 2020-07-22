const fs = require("fs");
const util = require("util");
const exists = util.promisify(fs.exists);
const unlink = util.promisify(fs.unlink);
const path = require("path");
const { homedir } = require("os");
const writeJSON = require("write-json-file");
const loadJSON = require("load-json-file");

const NHOST_DIR = path.join(homedir(), ".nhost");
const authPath = path.join(NHOST_DIR, "auth.json");

let configData = `# configurations used when starting your environment

# hasura graphql configuration
graphql_version: {{ hasura_gqe_version }}.cli-migrations
graphql_server_port: 8080
#graphql_admin_secret: (optional: if not specified, it will be generated on 'nhost dev')

# postgres configuration
postgres_version: {{ postgres_version }}
postgres_port: 5432
postgres_user: postgres
postgres_password: postgres
#postgres_db_data: (optional: if not specified, './db_data' will be used) 

# hasura backend plus configuration
backend_plus_version: {{ backend_version }}
backend_plus_port: 9000

# custom environment variables for Hasura GraphQL engine: webhooks, headers, etc
env_file: .env.development
`;

async function writeAuthFile(data) {
  try {
    return writeJSON.sync(authPath, data, {
      indent: 2,
      mode: 0o600,
    });
  } catch (err) {
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
  return loadJSON.sync(authPath);
}

async function authFileExists() {
  return await exists(authPath);
}

function getCustomApiEndpoint() {
  return "https://customapi.nhost.io";
}

function getNhostConfigTemplate() {
  return configData;
}

module.exports = {
  writeAuthFile,
  readAuthFile,
  authFileExists,
  getCustomApiEndpoint,
  getNhostConfigTemplate,
  deleteAuthFile,
};
