const dockerApiTemplate = `
FROM nhost/nodeapi:latest
WORKDIR /usr/src/app/
COPY package*.json yarn.lock ./
RUN yarn install
RUN yarn add express node-dir @babel/core @babel/cli @babel/preset-env @babel/polyfill @babel/plugin-transform-runtime @babel/node nodemon

# Unable to use COPY since we don't know if the API folder exists.
# So we need to juggle the folders a bit
RUN mkdir /usr/src/app/src/api
ADD . /tmp/app
RUN mkdir -p /tmp/app/api
RUN cp -r /tmp/app/api /usr/src/app/src/api
RUN rm -rf /tmp/app

# RUN ./node_modules/.bin/babel src -d dist
WORKDIR /usr/src/app/src
CMD ["../node_modules/.bin/nodemon", "--exec", "../node_modules/.bin/babel-node", "index.js"]
`;

function getDockerApiTemplate() {
  return dockerApiTemplate.trim();
}

module.exports = getDockerApiTemplate;
