const fetch = require("node-fetch");
const chalk = require("chalk");

const { readAuthFile, getCustomApiEndpoint } = require("./config");

async function validateAuth() {
  const url = getCustomApiEndpoint();
  const { email, token } = readAuthFile();

  let response;

  try {
    response = await fetch(`${url}/custom/cli/login/validate`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        email,
        token,
      }),
    });
  } catch (err) {
    throw new Error("Server unavailable. Please retry in a moment.");
  }

  const body = await response.json();
  if (!response.ok) {
    const { error = {} } = body;
    if (error.code === "server_not_available") {
      throw new Error("Server unavailable. Please retry in a moment.");
    }

    throw new Error("The provided token is not valid");
  }

  return body;
}

async function executeLogin(url, email) {
  let response;
  try {
    response = await fetch(`${url}/custom/cli/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        email,
      }),
    });
  } catch (err) {
    console.log(err);
    throw new Error("Server unavailable. Please retry in a moment.");
  }

  const body = await response.json();
  if (!response.ok) {
    const { error = {} } = body;
    if (error.code === "not_found") {
      throw new Error(
        `We couldn't find an account registered with ${chalk.bold.underline(
          email
        )}. Please register at https://nhost.io/register.`
      );
    }

    throw new Error(`Unexpected error: ${error}`);
  }

  return body;
}

module.exports = { executeLogin, validateAuth };
