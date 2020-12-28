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

The proxy sub-package provides two HTTP rate limited reverse proxy types:

- rateLimitedSingleRP: is a single host reverse proxy.
It embeds an httputil.ReverseProxy.

- rateLimitedMultipleRP: proxies requests to multiple hosts based on the request path.
It embeds an http.ServeMux which handlers are httputil.ReverseProxy. 

Both types needs to be initialized using the provied constructor. It enables the rate limiting functionality to be configured:
```Go
singleProxy := proxy.NewRateLimitedSingleRP(rate, urlToProxy)
```

```Go
multipleProxy := proxy.NewRateLimitedMultipleRP(rate, urlsToProxy...)
```

Rate limited is enforced at the struct level.
Therefore, for the multipleRP, a global rate limit is enforced whatever  backends host is targeted by the request.

Proxy configuration can be achieved by configuring the embedded structs:

- singleRP: httputil.ReverseProxy exposed as Server
```Go
singleProxy.Server.Transport = backend.Client().Transport
```
- multipleRP: http.ServeMux exposed as Router
```Go
multipleProxy.Router.Handle(customPattern, customHandler)
```
Examples are provided:
- [Single Host Reverse Proxy](examples/http-single-reverse-proxy/main.go)
- [Multiple Hosts Reverse Proxy](examples/http-multiple-reverse-proxy/main.go)


# Contributions

Contributions are welcomed :ok_hand:
