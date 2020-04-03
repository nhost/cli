const { Command, flags } = require("@oclif/command");
const process = require("process");
const { execSync } = require("child_process");
const fs = require("fs");

class DestroyCommand extends Command {
  async run() {
    this.log("shutting down all services...");
    execSync("docker-compose down > /dev/null 2>&1");

    const pidFile = "./.console.pid";
    if (!fs.existsSync(pidFile)) {
      this.log("nothing to do here");
      return;
    }

    const pid = fs.readFileSync(pidFile, { encoding: "utf8" });
    process.kill(parseInt(pid), "SIGINT");

    fs.unlinkSync(pidFile);
    fs.unlinkSync("./docker-compose.yaml");
    this.log("done.");
  }
}

DestroyCommand.description = `Describe the command here
...
Extra documentation goes here
`;

DestroyCommand.flags = {
  name: flags.string({ char: "n", description: "name to print" })
};

module.exports = DestroyCommand;
