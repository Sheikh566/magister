## Connection Keep-Alive

In older versions of HTTP (HTTP/1.0 without keep-alive), a client would open a new TCP connection for every single request. This was incredibly slow and resource-intensive because the TCP handshake takes time.

HTTP/1.1 changed the default behavior: **connections are persistent by default**. This means a single TCP connection can be used to send multiple requests and receive multiple responses sequentially.

### The Request-Response Cycle

To support persistent connections, your server cannot simply close the socket after sending a response. Instead, it must:
1. Read a request.
2. Send the full response (including body).
3. Immediately go back to step 1 and wait for another request on the same socket.

This highlights why `Content-Length` is so critical: the client relies on it to know when one response ends so it can start reading the next one!

### External Resources
* [MDN: Connection Management in HTTP/1.x](https://developer.mozilla.org/en-US/docs/Web/HTTP/Connection_management_in_HTTP_1.x)
* [RFC 9112: Persistent Connections](https://httpwg.org/specs/rfc9112.html#persistent.connections)

### Your Task
Update your server to support persistent connections. After serving a request, do not close the socket. The tester will send a request, read your response, and then immediately send a second request on the same open TCP connection. You must successfully handle and respond to both.
