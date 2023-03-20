# Capy content storage
![capy-content-storage](https://github.com/krol44/capy-content-storage/blob/master/.github/readme-capybara.gif?raw=true)

### **It's file storage which working through api**

## Install:
1. create dirs ```mkdir /tmp/files && mkdir /tmp/files-removed```
2. run
```
docker run -d \
    -p 127.0.0.1:8017:8017 \
    --mount type=bind,source=/tmp/files,target=/files \
    --mount type=bind,source=/tmp/files-removed,target=/files-removed \
    -e DEV="false" \
    -e HOST_URL="localhost" \
    -e LIMIT_UPLOAD_MB="100" \
    -e TOKEN="some-token" \
    -e TZ="Europe/Moscow" \
    --restart=always --log-opt max-size=5m --name=capy-content-storage \
    krol44/capy-content-storage
```

or

1. git clone
2. mkdir /tmp/files && mkdir /tmp/files-removed
3. ./docker-start.sh
4. ./test-request.sh

## Api:
### Save file:
```Storage``` - key space / example: *site1.test, site2, photos, videos, etc*
```
curl -H "Token: sogime-token" -H "Storage: langlija" \
 -F "file=@test-img.png" --insecure http://localhost:8017/upload
```
Result:
```
{
    "status":true,
    "hostUrl":"localhost",
    "pathServer":"langlija/c/07/c07441a2d2c1dcbde7908dfa96625cc0.png",
    "size":3933042,
    "filenameUploaded":"test-img.png"
}
```

### Get files:
```-d '{"paths": [""]}'``` - get only stats

or

```-d '{"paths": ["all"]}'``` - get all files (heavy request)

or

```
curl -H "Token: some-token" \
    -H "Storage: langlija" --insecure http://localhost:8017/files \
    -H "Content-Type: application/json" \
    -d '{"paths": ["langlija/e/31/e3124fd2ae8e4bdf5793b46d694052d1.png","langlija/c/f3/cf364e84cbdc6416cdc80d8ef6bb10db.png"]}'
```

Result:
```
{
    "status":true,
    "quantity":15,
    "itemsRemoved":0,
    "size":58995630,
    "sizeRemoved":0,
    "Files":[
        {
            "filename":"langlija/c/f3/cf364e84cbdc6416cdc80d8ef6bb10db.png",
            "size":3933042,
            "modTime":"2023-03-19T18:48:01.430306806+03:00"
        },
        {
            "filename":"langlija/e/31/e3124fd2ae8e4bdf5793b46d694052d1.png",
            "size":3933042,
            "modTime":"2023-03-19T18:48:06.172489335+03:00"
        }
    ]
}
```

### Remove file:
```
curl -H "Token: some-token" \
    -H "Storage: langlija" --insecure http://localhost:8017/remove \
    -H "Content-Type: application/json" \
    -d '{"path":"langlija/e/31/e3124fd2ae8e4bdf5793b46d694052d1.png"}'
```
Result:

file will be copy the dir (files-removed) and removed from the dir (files)
```
{
    "status":true
}
```

### Error json type
```
{
    "status":false,
    "message":"error removing",
    "error":"open files/langlija/e/31/e3124fd2ae8e4bdf5793b46d694052d1.png: no such file or directory"
}
```