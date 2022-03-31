# fo-aas

fo-aas provide a proxy server to foaas, by connecting and fetching its response. It enforces user based throttling for incomming requests and provides basic RED metrics.

## Building and running the app

The app expects the following flags

- `host` : listen interface. (def: `"0.0.0.0"`)
- `port` : port to listen on. (def: `8080`)
- `rate-limit` : limit to enforce per user, expressed as requests per second. (def: `10` requests per second)

### using docker

```bash
make build-docker
make run
```

Sample execution

```bash
curl 'http://localhost:8080/message' -H "User: pable"
```

## Metrics

Local [grafana]( http://localhost:3000/ ) service can be accessed to

Credentias:

- user: `admin`
- passwd: `foobar`
