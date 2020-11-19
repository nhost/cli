const dockerApiTemplate = `
FROM nodeapi:latest
WORKDIR /usr/src/app/
COPY package*.json yarn.lock ./
RUN yarn install
RUN yarn add express node-dir
RUN yarn add -D @babel/core @babel/cli @babel/preset-env @babel/plugin-transform-runtime
COPY api src/api
RUN ./node_modules/.bin/babel src -d dist
WORKDIR /usr/src/app/dist
CMD ["node", "index.js"]
`;

function getDockerApiTemplate() {
  return dockerApiTemplate.trim();
}

module.exports = getDockerApiTemplate;
