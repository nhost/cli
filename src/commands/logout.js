const { Command } = require("@oclif/command");
const { authFileExists } = require("../util/config");
const chalk = require("chalk");
const { deleteAuthFile } = require("../util/config");

class LogoutCommand extends Command {
  async run() {
    // check if auth file exists
    if (!(await authFileExists())) {
      this.log(
        `${chalk.red(
          "No credentials found!"
        )} Please login first with ${chalk.bold.underline("nhost login")}`
      );
      this.exit(1);
    }

    try {
      await deleteAuthFile();
    } catch (err) {
      this.log(`${chalk.red("Error!")} ${err.message}`);
    }

    this.log(`\n${chalk.white("logged out")}`);
  }
}

LogoutCommand.description = `Logout from your Nhost account
...
Logout from your Nhost account
`;

module.exports = LogoutCommand;
