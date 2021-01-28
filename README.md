# Limit request count service

## Prerequisite
docker install on Mac
```bash
https://hub.docker.com/editions/community/docker-ce-desktop-mac/
```

## Start service
1. use docker-compose

```bash
$ docker-compose up
```

2. in another terminal to send request and you can get the response that nth request number
```bash
$ curl http://localhost:8080/page
```

## Test
```bash
$ go test -v
```