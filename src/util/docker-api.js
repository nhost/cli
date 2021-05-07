const dockerApiTemplate = `
FROM nhost/nodeapi:v0.2.7
WORKDIR /usr/src/app

COPY api ./api

RUN ./install.sh

ENTRYPOINT ["./entrypoint-dev.sh"]
`;
function getDockerApiTemplate() {
  return dockerApiTemplate.trim();
}

module.exports = getDockerApiTemplate;
