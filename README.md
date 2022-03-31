# fo-aas

fo-aas provide a proxy server to foaas, by connecting and fetching its response. It enforces user based throttling for incomming requests and provides basic RED metrics.

Authorization was out the scope of this implementation, so the applicaiton expects a specific HTTP header `"User"` with the user identifier, otherwise the request es reject.

The implementation of the rate limiting is session based. Every user has a quota of tokens (requests), that can be consumed within a session. Sessions have fixed length and on expiration, tokens are re-allocated.

## Building and running the app

The app expects the following flags

- `host` : listen interface. (def: `"0.0.0.0"`)
- `port` : port to listen on. (def: `8080`)
- `session-length` : amount of time until the session expires in milliseconds  (def: `1000` requests per second)
- `tokens` : amount of tokens available per user by session (def: `1000`)
- `timeout` : timeout in milleconds to establish a connection with foaas (def: `3000` m sec)

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

Metrics are published by the app on the `/metrics` endpoint and collected by the prometheus service running on docker.

Local [grafana]( http://localhost:3000/ ) service can be accessed to check metrics. There's an already simple dashboard built that can be imported under `dev/grafana/provisioning/dashboards/metrics.json` called **Metrics**.

Credentias:

- user: `admin`
- passwd: `foobar`
