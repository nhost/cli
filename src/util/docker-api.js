const dockerApiTemplate = `
FROM nhost/nodeapi:v0.1.4
WORKDIR /usr/src/app

COPY api ./api

RUN ./install.sh

CMD ["./node_modules/.bin/nodemon", "-L", "--watch", "api", "--exec", "./node_modules/.bin/babel-node", "index.js"]
`;
function getDockerApiTemplate() {
  return dockerApiTemplate.trim();
}

module.exports = getDockerApiTemplate;
