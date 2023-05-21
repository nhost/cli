# Troubleshooting

When things don't go as expected the CLI will not stop the containers automatically, for instance:


```
$ nhost dev up
Verifying configuration...
Configuration is valid!
Setting up Nhost development environment...
Starting Nhost development environment...
[+] Running 10/10
 ✔ Container existingproject-traefik-1    Running                                                      0.0s
 ✔ Container existingproject-minio-1      Running                                                      0.0s
 ✘ Container existingproject-postgres-1   Error                                                        0.5s
 ✔ Container existingproject-mailhog-1    Running                                                      0.0s
 ✔ Container existingproject-dashboard-1  Running                                                      0.0s
 ✔ Container existingproject-functions-1  Healthy                                                      0.5s
 ✔ Container existingproject-graphql-1    Running                                                      0.0s
 ✔ Container existingproject-storage-1    Running                                                      0.0s
 ✔ Container existingproject-console-1    Running                                                      0.0s
 ✔ Container existingproject-auth-1       Running                                                      0.0s
dependency failed to start: container existingproject-postgres-1 is unhealthy
failed to start Nhost development environment: failed to start docker compose: exit status 1
- Do you want to stop Nhost development environment it? [y/N]
```

While the environment is broken but hasn't been stopped you can check the logs with `nhost dev logs` or run any arbitrary `docker compose` (already configured for your project) with `nhost dev compose ...`

```
$ nhost dev logs graphql
...
existingproject-graphql-1  | {"detail":{"http_info":{"content_encoding":null,"http_version":"HTTP/1.1","ip":"192.168.128.9","method":"POST","status":400,"url":"/v1/metadata"},"operation":{"error":{"code":"already-tracked","error":"view/table already tracked: \"storage.buckets\"","path":"$.args"},"query":{"type":"pg_track_table"},"request_id":"a3e88205-8e48-4075-b4ca-d68c97e71d40","request_mode":"error","response_size":100,"uncompressed_response_size":100,"user_vars":{"x-hasura-role":"admin"}},"request_id":"a3e88205-8e48-4075-b4ca-d68c97e71d40"},"level":"error","timestamp":"2023-05-20T17:10:01.451+0000","type":"http-log"}
existingproject-graphql-1  | {"detail":{"http_info":{"content_encoding":null,"http_version":"HTTP/1.1","ip":"192.168.128.9","method":"POST","status":400,"url":"/v1/metadata"},"operation":{"error":{"code":"already-tracked","error":"view/table already tracked: \"storage.files\"","path":"$.args"},"query":{"type":"pg_track_table"},"request_id":"50140ff3-5fcf-494b-8178-dafda75e774d","request_mode":"error","response_size":98,"uncompressed_response_size":98,"user_vars":{"x-hasura-role":"admin"}},"request_id":"50140ff3-5fcf-494b-8178-dafda75e774d"},"level":"error","timestamp":"2023-05-20T17:10:01.451+0000","type":"http-log"}
...

$ nhost dev compose top
existingproject-auth-1
PID    USER   TIME   COMMAND
6122   root   0:00   node /usr/local/bin/pnpm run start
6289   root   0:01   node ./dist/start.js

existingproject-console-1
PID    USER   TIME   COMMAND
6120   root   0:00   hasura-cli console --no-browser --endpoint http://graphql:8080 --address 0.0.0.0 --console-port 9695 --api-port 443 --api-host https://local.hasura.nhost.run --console-hge-endpoint https://local.hasura.nhost.run

existingproject-dashboard-1
PID    USER   TIME   COMMAND
2330   1001   0:00   node dashboard/server.js
...
```
