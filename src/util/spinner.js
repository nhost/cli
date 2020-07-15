const ora = require("ora");
const chalk = require("chalk");

module.exports = function (message, delay = 300) {
  let spinner;
  let running = false;

  const planned = setTimeout(() => {
    spinner = ora(chalk.green(message));
    spinner.color = "green";
    spinner.start();
    running = true;
  }, delay);

  const cancel = () => {
    clearTimeout(planned);

    if (running) {
      spinner.stop();
      running = false;
    }

    process.removeListener("nhostExit", cancel);
  };

  process.on("nhostExit", cancel);
  return cancel;
};
