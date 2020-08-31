const inquirer = require("inquirer");

module.exports = async function (projects) {
  const user_projects = projects.filter((project) => project.user_id);
  const team_projects = projects.filter((project) => project.team_id);

  const choices = [];
  if (user_projects.length > 0) {
    choices.push(new inquirer.Separator("---personal---"));
  }
  choices.push(
    ...user_projects.map((project) => {
      return { name: project.name, value: project.id };
    })
  );

  const grouped_team_projects = team_projects.reduce((r, project) => {
    r[project.team.name] = [...(r[project.team.name] || []), project];
    return r;
  }, {});

  for (const [key, value] of Object.entries(grouped_team_projects)) {
    choices.push(new inquirer.Separator(`---team ${key}---`));
    choices.push(
      ...value.map((m) => {
        return { name: m.name, value: m.id };
      })
    );
  }

  const selectedProject = await inquirer.prompt({
    type: "list",
    name: "id",
    message: "Select Project",
    choices,
  });

  return selectedProject.id;
};
