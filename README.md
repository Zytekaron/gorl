# gorl

**version:** v0.1.0

gorl is:
- A thread-safe implementation of the [Leaky Bucket](https://en.wikipedia.org/wiki/Leaky_bucket) rate limiting algorithm
- A server-side library to prevent clients from sending excess requests to your server

gorl is not:
- A client-side rate limiting solution to block draw attempts until tokens have been refilled.
  Try [uber-go/ratelimit](https://github.com/uber-go/ratelimit) instead!

## Usage

```go
func RateLimit(bm *gorl.BucketManager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        b := bm.Get(getIP(r))
        if !b.Draw(1) { // prevent requests if drawing tokens failed
            w.WriteHeader(429)
            w.Write([]byte("Too Many Requests"))
            return
        }
        // otherwise, good to go!
        w.WriteHeader(200)
        w.Write([]byte("Success!"))
    }
}
```

Note: If you plan on using this to prevent repeated attempts to authenticate
to a server with invalid credentials, be sure to get the draw logic correct!

**recommended control logic:** draw tokens *only* when authentication fails...
*there is an issue:* a malicious user may send a large quantity of requests at
one instant and have them all report a successful CanDraw at once, before the
server has a chance to validate the other ongoing request auths and draw tokens.
This issue is mitigated by using ForceDraw when authentication fails, because
if a malicious client was able to successfully complete more authentication
attempts in one refill period than is typically allowed, they will have a
negative quantity of tokens for the subsequent time periods, averaging it out.

Code sample for the above control logic:
```go
b := bm.Get(getIP(req))
// prevent requests from IPs which have exceeded the limit (do not attempt to draw here)
if !b.Draw(1) {
    return // 429 here
}
// prevent requests with invalid authentication (force draw here if auth fails!)
if !hasValidAuth(r) {
    b.ForceDraw(1)
    return // 401/403 here
}
// continue handling the request (do not draw here).
// if the authentication was successful, the user should not be limited
// by this middleware, but instead another (authed-level) one if needed.
...
```

## License
**gorl** is licenced under the [MIT License](./LICENSE).
