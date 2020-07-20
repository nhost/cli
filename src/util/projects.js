const inquirer = require("inquirer");

module.exports = async function (projects) { 
  const choices = projects.map((project) => {
    return { name: project.name, value: project.id };
  });

  const selectedProject = await inquirer
    .prompt({
      type: "list",
      name: "id",
      message: "Select Project",
      choices,
    })

  return selectedProject.id
}