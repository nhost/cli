const dockerApiTemplate = `
FROM nodeapi:latest

COPY package*.json yarn.lock ./
COPY api src/api
RUN yarn install
RUN ./node_modules/.bin/babel src -d dist
WORKDIR /usr/src/app/dist

CMD ["node", "index.js"]`;

function getDockerApiTemplate() {
  return dockerApiTemplate;
}

module.exports = getDockerApiTemplate;
