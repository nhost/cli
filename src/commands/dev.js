const { Command, flags } = require('@oclif/command')
const fs = require ('fs');
const { execSync } = require('child_process');
const yaml = require('js-yaml');
const nunjucks = require('nunjucks');

class DevCommand extends Command {
  async run() {
    if (!fs.existsSync('./docker-compose.example') || !fs.existsSync('./config.yaml')) {
      return this.log('Please run `nhost-cli init` before starting a development environment.');
    }

    const nhostConfig = yaml.safeLoad(fs.readFileSync('config.yaml', { encoding: 'utf8' }));
    const dockerComposeTemplate = fs.readFileSync('docker-compose.example', { encoding: 'utf8' });

    fs.writeFileSync('docker-compose.yaml', nunjucks.renderString(dockerComposeTemplate, nhostConfig));
    execSync('docker-compose up -d');
    this.log('services are launching...');
    
    // check whether the graphql-engine endpoint is up & running
    let reachable = false;
    while (!reachable) {
      try {
        execSync(`hasura console --endpoint=http://localhost:${nhostConfig.graphql_server_port} --admin-secret=${nhostConfig.graphql_admin_secret}`);
      } catch (Error) {
        continue;
      }
      reachable = true;
    }
  }
}

DevCommand.description = `Describe the command here
...
Extra documentation goes here
`

DevCommand.flags = {
  name: flags.string({char: 'n', description: 'name to print'}),
}

module.exports = DevCommand

nunjucks.configure({ autoescape: true });