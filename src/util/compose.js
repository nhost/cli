function generateNhostBackendYaml(options) {
  const {
    postgres_image,
    postgres_version,
    postgres_port,
    postgres_user,
    postgres_password,
    hasura_graphql_engine,
    hasura_graphql_version,
    hasura_graphql_port,
    hasura_graphql_admin_secret,
    hasura_backend_plus_version,
    hasura_backend_plus_port,
    graphql_jwt_key,
    api_port,
    minio_port,
    provider_success_redirect,
    provider_failure_redirect,
    google_enable,
    google_client_id,
    google_client_secret,
    github_enable,
    github_client_id,
    github_client_secret,
    facebook_enable,
    facebook_client_id,
    facebook_client_secret,
    linkedin_enable,
    linkedin_client_id,
    linkedin_client_secret,
    env_file,
    startAPI,
  } = options;

  const hasuraGraphQLEngine = hasura_graphql_engine
    ? hasura_graphql_engine
    : "hasura/graphql-engine";

  const postgresImage = postgres_image
    ? postgres_image
    : `postgres:${postgres_version}`;

  const project = {
    version: "3.6",
    services: {
      ["nhost-postgres"]: {
        container_name: "nhost_postgres",
        image: postgresImage,
        ports: [`${postgres_port}:5432`],
        restart: "always",
        environment: {
          POSTGRES_USER: postgres_user,
          POSTGRES_PASSWORD: postgres_password,
        },
        volumes: [`./db_data:/var/lib/postgresql/data`],
      },
      [`nhost-graphql-engine`]: {
        container_name: "nhost_hasura",
        image: `${hasuraGraphQLEngine}:${hasura_graphql_version}`,
        ports: [`${hasura_graphql_port}:${hasura_graphql_port}`],
        depends_on: [`nhost-postgres`],
        restart: "always",
        environment: {
          HASURA_GRAPHQL_SERVER_PORT: hasura_graphql_port,
          HASURA_GRAPHQL_DATABASE_URL: `postgres://${postgres_user}:${postgres_password}@nhost-postgres:5432/postgres`,
          HASURA_GRAPHQL_ENABLE_CONSOLE: "false",
          HASURA_GRAPHQL_ENABLED_LOG_TYPES: `startup, http-log, webhook-log, websocket-log, query-log`,
          HASURA_GRAPHQL_ADMIN_SECRET: hasura_graphql_admin_secret,
          HASURA_GRAPHQL_JWT_SECRET: `{"type":"HS256", "key": "${graphql_jwt_key}"}`,
          HASURA_GRAPHQL_MIGRATIONS_SERVER_TIMEOUT: 20,
          HASURA_GRAPHQL_NO_OF_RETRIES: 20,
          HASURA_GRAPHQL_UNAUTHORIZED_ROLE: "public",
          NHOST_HASURA_URL: `http://nhost_hasura:${hasura_graphql_port}/v1/graphql`,
          NHOST_WEBHOOK_SECRET: `devnhostwebhooksecret`,
          NHOST_HBP_URL: `http://nhost_hbp:${hasura_backend_plus_port}`,
          NHOST_CUSTOM_API_URL: `http://nhost_api:${api_port}`,
        },
        env_file: [env_file],
        command: ["graphql-engine", "serve"],
        volumes: ["../nhost/migrations:/hasura-migrations"],
      },
      [`nhost-hasura-backend-plus`]: {
        container_name: "nhost_hbp",
        image: `nhost/hasura-backend-plus:${hasura_backend_plus_version}`,
        ports: [`${hasura_backend_plus_port}:${hasura_backend_plus_port}`],
        depends_on: [`nhost-graphql-engine`],
        restart: "always",
        environment: {
          PORT: hasura_backend_plus_port,
          USER_FIELDS: "",
          USER_REGISTRATION_AUTO_ACTIVE: "true",
          DATABASE_URL: `postgres://${postgres_user}:${postgres_password}@nhost-postgres:5432/postgres`,
          HASURA_GRAPHQL_ENDPOINT: `http://nhost-graphql-engine:${hasura_graphql_port}/v1/graphql`,
          HASURA_ENDPOINT: `http://nhost-graphql-engine:${hasura_graphql_port}/v1/graphql`,
          HASURA_GRAPHQL_ADMIN_SECRET: hasura_graphql_admin_secret,
          JWT_ALGORITHM: "HS256",
          JWT_KEY: graphql_jwt_key,
          AUTH_ACTIVE: "true",
          AUTH_LOCAL_ACTIVE: "true",
          REFRESH_TOKEN_EXPIRES: 43200,
          JWT_TOKEN_EXPIRES: 15,
          S3_ENDPOINT: `nhost_minio:${minio_port}`,
          S3_SSL_ENABLED: "false",
          S3_BUCKET: "nhost",
          S3_ACCESS_KEY_ID: "minioaccesskey123123",
          S3_SECRET_ACCESS_KEY: "miniosecretkey123123",
          LOST_PASSWORD_ENABLE: "true",
          PROVIDER_SUCCESS_REDIRECT: provider_success_redirect,
          PROVIDER_FAILURE_REDIRECT: provider_failure_redirect,
          GOOGLE_ENABLE: google_enable.toString(),
          GOOGLE_CLIENT_ID: google_client_id,
          GOOGLE_CLIENT_SECRET: google_client_secret,
          GITHUB_ENABLE: github_enable.toString(),
          GITHUB_CLIENT_ID: github_client_id,
          GITHUB_CLIENT_SECRET: github_client_secret,
          FACEBOOK_ENABLE: facebook_enable.toString(),
          FACEBOOK_CLIENT_ID: facebook_client_id,
          FACEBOOK_CLIENT_SECRET: facebook_client_secret,
          LINKEDIN_ENABLE: linkedin_enable.toString(),
          LINKEDIN_CLIENT_ID: linkedin_client_id,
          LINKEDIN_CLIENT_SECRET: linkedin_client_secret,
        },
        env_file: [env_file],
        volumes: ["../nhost/custom:/app/custom"],
      },
      [`minio`]: {
        container_name: "nhost_minio",
        image: "minio/minio",
        user: "999:1001",
        restart: "always",
        volumes: ["./minio/data:/data", "./minio/config:/.minio"],
        environment: {
          MINIO_ACCESS_KEY: "minioaccesskey123123",
          MINIO_SECRET_KEY: "miniosecretkey123123",
        },
        entrypoint: "sh",
        command: `-c 'mkdir -p /data/nhost && /usr/bin/minio server --address :${minio_port} /data'`,
        ports: [`${minio_port}:${minio_port}`],
      },
    },
  };

  if (startAPI) {
    project["services"]["nhost-api"] = {
      container_name: "nhost_api",
      build: {
        context: "../",
        dockerfile: "./.nhost/Dockerfile-api",
      },
      environment: {
        PORT: api_port,
        NHOST_JWT_ALGORITHM: "HS256",
        NHOST_JWT_KEY: graphql_jwt_key,
        NHOST_HASURA_URL: `http://nhost_hasura:${hasura_graphql_port}/v1/graphql`,
        NHOST_HASURA_ADMIN_SECRET: hasura_graphql_admin_secret,
        NHOST_WEBHOOK_SECRET: "devnhostwebhooksecret",
        NHOST_HBP_URL: `http://nhost_hbp:${hasura_backend_plus_port}`,
        NHOST_CUSTOM_API_URL: `http://nhost_api:${api_port}`,
      },
      ports: [`${api_port}:${api_port}`],
      env_file: [env_file],
      volumes: ["../api:/usr/src/app/api"],
    };
  }

  return project;
}

module.exports = {
  generateNhostBackendYaml,
};
