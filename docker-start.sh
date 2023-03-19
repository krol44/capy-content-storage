#!/bin/bash
docker build . -t capy-content-storage-image
docker rm -f capy-content-storage
docker run -d -e TZ="Europe/Moscow" \
  -p 0.0.0.0:8017:8017 \
  --mount type=bind,source=/tmp/files,target=/files \
  --mount type=bind,source=/tmp/files-removed,target=/files-removed \
  -e DEV="FALSE" \
  -e HOST_URL="localhost" \
  -e LIMIT_UPLOAD_MB="100" \
  -e TOKEN="some-token" \
  --restart=always --log-opt max-size=5m --name=capy-content-storage capy-content-storage-image