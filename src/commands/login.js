const { Command, flags } = require("@oclif/command");
const { validate } = require("email-validator");
const promptEmail = require("email-prompt");
const stringify = require("querystring");
const fetch = require("node-fetch");
const chalk = require("chalk");

const spinner = require("../util/spinner");
const executeLogin = require("../util/login");
const sleep = require("../util/sleep");
const writeAuthFile = require("../util/config");

class LoginCommand extends Command {
  async readEmail() {
    let email;

    try {
      email = await promptEmail({ start: "Enter your email: " });
    } catch (err) {
      this.log(); // \n

      if (err.message === "User abort") {
        throw new Error(`${chalk.red("Aborted!")}`);
      }

      if (err.message === "stdin lacks setRawMode support") {
        throw new Error(
          error(
            "Interactive mode not supported – please run login --email=email"
          )
        );
      }
    }

    this.log(); // \n
    return email;
  }

  async verify({ apiUrl, email, verificationToken }) {
    const query = {
      email,
      token: verificationToken,
    };

    let res;
    try {
      res = await fetch(
        `${apiUrl}/custom/cli/login/verify?${stringify.stringify(query)}`
      );
    } catch (err) {
      throw new Error(
        `An unexpected error occurred when trying to verify your login: ${err.message}`
      );
    }

    let body;
    try {
      body = await res.json();
    } catch (err) {
      throw new Error(
        `An unexpected error occurred when trying to verify your login: ${err.message}`
      );
    }

    return body;
  }

  async run() {
    const { flags } = this.parse(LoginCommand);
    let email = flags.email;
    // const apiUrl = "https://customapi.nhost.io";
    const apiUrl = "http://localhost:3006";
    let emailIsValid = false;
    let stopSpinner;

    // if email was passed as an argument with --email=email
    if (email) {
      if (!validate(email)) {
        this.log(`${chalk.red("Invalid email:")} ${email}.`);
        this.exit(1);
      }
    } else {
      do {
        try {
          email = await this.readEmail();
        } catch (err) {
          this.log(err.message);
          this.exit(1);
        }

        emailIsValid = validate(email);
        if (!emailIsValid) {
          this.log(`${chalk.red("Invalid email:")} ${email}.`);
        }
      } while (!emailIsValid);
    }

    stopSpinner = spinner("An email is being sent to you.");

    let verificationToken;
    try {
      const loginResponse = await executeLogin(apiUrl, email);
      verificationToken = loginResponse.verificationToken;
    } catch (err) {
      stopSpinner();
      this.log(`${chalk.red("Error!")} ${err.message}`);
      this.exit(1);
    }

    stopSpinner();

    // Clear up `Sending email` success message
    // process.stdout.write(eraseLines(possibleAddress ? 1 : 2));

    this.log(
      `An email was sent to ${chalk.bold.underline(
        email
      )}. Please follow the instructions inside it.`
    );
    stopSpinner = spinner("Waiting for your confirmation.");

    let token;
    while (!token) {
      try {
        await sleep(1000); // 1 second
        token = await this.verify({ apiUrl, email, verificationToken });
      } catch (err) {
        if (/invalid json response body/.test(err.message)) {
          // /now/registraton is currently returning plain text in that case
          // we just wait for the user to click on the link
        } else {
          stopSpinner();
          console.log(err.message);
          this.exit(1);
        }
      }
    }

    stopSpinner();
    this.log(`${chalk.cyan("Email Confirmed!")} You are now logged in.`);

    // write auth.json to the user's home directoy
    writeAuthFile({ email, token: token.token });
  }
}

LoginCommand.description = `
`;

LoginCommand.flags = {
  email: flags.string({}),
};

module.exports = LoginCommand;
