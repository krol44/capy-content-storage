#!/bin/bash

printf "\n\n"
curl -H "Token: some-token" -H "Storage: langlija" -F "file=@test-img.png" --insecure http://localhost:8017

printf "\n\n"
curl -H "Token: some-token" \
    -H "Storage: langlija" --insecure http://localhost:8017/files \
    -H "Content-Type: application/json" \
    -d '{"paths": ["langlija/e/31/e3124fd2ae8e4bdf5793b46d694052d1.png","langlija/c/f3/cf364e84cbdc6416cdc80d8ef6bb10db.png"]}'

printf "\n\n"
curl -H "Token: some-token" \
    -H "Storage: langlija" --insecure http://localhost:8017/files \
    -H "Content-Type: application/json" \
    -d '{"paths": []}'

#printf "\n\n"
#curl -H "Token: some-token" \
#    -H "Storage: langlija" --insecure http://localhost:8017/files \
#    -H "Content-Type: application/json" \
#    -d '{"paths": ["all"]}'

printf "\n\n"
curl -H "Token: some-token" \
    -H "Storage: langlija" --insecure http://localhost:8017/remove \
    -H "Content-Type: application/json" \
    -d '{"path":"langlija/e/31/e3124fd2ae8e4bdf5793b46d694052d1.png"}'

printf "\n\n"