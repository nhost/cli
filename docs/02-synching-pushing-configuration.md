# Synching and Pushing Configuration

1. Pull configuration from the cloud

```
$ nhost config pull
- /path/to/existingproject/nhost/nhost.toml already exists. Do you want to overwrite it? [y/N] y
- /path/to/existingproject/.secrets already exists. Do you want to overwrite it? [y/N] y
Pulling config from Nhost...
Getting secrets list from Nhost...
Adding .secrets to .gitignore...
Success!
- Review `nhost/nhost.toml` and make sure there are no secrets before you commit it to git.
- Review `.secrets` file and set your development secrets
- Review `.secrets` was added to .gitignore
```

2. The previous command basically downloaded the configuration from the cloud and overwrote the local changes so now you need to review the changes and make sure they are what you want:

```
❯ git diff nhost/nhost.toml
diff --git a/nhost/nhost.toml b/nhost/nhost.toml
index 623504f..cccf89e 100644
 [hasura]
-version = 'v2.25.0-ce'
+version = 'v2.15.2'
 adminSecret = '{{ secrets.HASURA_GRAPHQL_ADMIN_SECRET }}'
 webhookSecret = '{{ secrets.NHOST_WEBHOOK_SECRET }}'

@@ -10,25 +17,14 @@ type = 'HS256'
 key = '{{ secrets.HASURA_GRAPHQL_JWT_SECRET }}'

 [hasura.settings]
-corsDomain = ['*']
-devMode = true
-enableAllowList = false
-enableConsole = true
 enableRemoteSchemaPermissions = false
-enabledAPIs = ['metadata', 'graphql', 'pgdump', 'config']
-
-[hasura.logs]
-level = 'warn'
-
-[hasura.events]
-httpPoolSize = 100

 [functions]
 [functions.node]
 version = 16

```

In the previous output we can see that the hasura version running in the cloud is different from the local development environment and that there are a few new supported settings for hasura.

IMPORTANT: If your repo contains the file `nhost/nhost.toml` we will replace your cloud configuration with the configuration specified in that file.
