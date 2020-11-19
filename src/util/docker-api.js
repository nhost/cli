const dockerApiTemplate = `
FROM nodeapi:latest
WORKDIR /usr/src/app/
COPY package*.json yarn.lock ./
RUN yarn install
RUN yarn add express node-dir @babel/core @babel/cli @babel/preset-env @babel/polyfill @babel/plugin-transform-runtime @babel/node nodemon
COPY api src/api
# RUN ./node_modules/.bin/babel src -d dist
WORKDIR /usr/src/app/src
CMD ["../node_modules/.bin/nodemon", "--exec", "../node_modules/.bin/babel-node", "index.js"]
`;

function getDockerApiTemplate() {
  return dockerApiTemplate.trim();
}

module.exports = getDockerApiTemplate;
