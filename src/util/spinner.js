const ora = require("ora");
const chalk = require("chalk");

module.exports = function (message) {
  console.log(); // \n

  let spinner = ora(chalk.white(message))
  spinner.color = "white";
  spinner.start();
  let running = true;

  const cancel = () => {
    if (running) {
      spinner.stop();
      running = false;
    }
  };

  return {spinner: spinner, stopSpinner: cancel};
};

