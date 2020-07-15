const fetch = require("node-fetch");
const chalk = require("chalk");

module.exports = async function (url, email) {
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
        `We couldn't find an account registered with ${chalk.bold.underline(email)}. Please register at https://nhost.io/register.`
      );
    }

    throw new Error(`Unexpected error: ${error}`);
  }

  return body;
};
