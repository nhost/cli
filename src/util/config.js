
const path = require("path");
const { homedir } = require("os");
const writeJSON = require("write-json-file");

const NHOST_DIR = path.join(homedir(), ".nhost");
const authPath = path.join(NHOST_DIR, 'auth.json');

module.exports = function (data) {
  try {
    return writeJSON.sync(authPath, data, {
      indent: 2,
      mode: 0o600,
    });
  } catch (err) {
    throw err;
  }
};
