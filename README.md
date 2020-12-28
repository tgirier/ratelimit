![rate-limit](https://i.imgur.com/QN5jTS7.gif)

```Go
import "github/tgirier/ratelimit"
```

Rate Limit is a Go API providing rate limited functionalities.

You can choose from to various types according to your needs.
Each type comes with a set of rate-limited methods.

Built with :heart: in Paris.

:raised_hands: A huge thanks to @bitfield for his mentoring on this project. :raised_hands:

# Installation

To import the API within your Go program, simply add the following statement to your go package:
```Go
import "github/tgirier/ratelimit"
```
or (if you want to you use the proxy sub-package):
```Go
import "github.com/tgirier/ratelimit/proxy"
```

# Usage

Rate Limit regroups two packages:
- Main rate limit package
- Proxy sub package

## Main package

Two rate limited types are available within the main package:
- **Worker (general purpose)**: executes a given function at a given rate.
- **HTTPClient (http only)**: executes HTTP requests at a given rate

Both types need to be initialized using the corresponding constructor. It enables the rate limiting functionality to be configured:
```Go
c := ratelimit.NewHTTPClient(rate)
```

The HTTPClient embeds an http.Client.

Examples are provided:
- [General rate limiter](examples/general-rate-limit/main.go)
- [HTTP client](examples/http-client/main.go)

## Proxy package

The proxy sub-package provides an HTTP rate limited reverse proxy type.
An httputil.ReverseProxy is embedded with the reverse proxy type.

The proxy type needs to be initialized using the provied constructor. It enables the rate limiting functionality to be configured:
```Go
proxy := proxy.NewRateLimitedReverseProxy(urlToProxy, rate)
```

All the requests will be proxied to the given URL at a given rate.
Proxy configuration can be achieved by configuring the embedded httputil.ReverseProxy (embedded as Server):
```Go
proxy.Server.Transport = backend.Client().Transport
```

A detailed example is provided [here](examples/http-single-reverse-proxy/main.go).

# Contributions

Contributions are welcomed :ok_hand:
