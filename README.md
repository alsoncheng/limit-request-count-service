# Limit request count service

1. start the service with docker-compose

```bash
$ docker-compose up
```

2. send request and you can get the response that nth request number
```bash
$ curl http://localhost:8080/page
```

## Test
```bash
$ go test -v
```