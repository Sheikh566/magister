## Finding the Right Path

A web server rarely serves just a single resource. It must inspect the incoming HTTP Request to figure out what the client is asking for. The primary way it does this is by looking at the **Request Target** (often just called the path or URI).

The Request Line of an HTTP message looks like this:
```
METHOD /path/to/resource HTTP/VERSION
```

For example: `GET /index.html HTTP/1.1`.

### Status Codes Matter

HTTP uses Status Codes to communicate the outcome of a request:
* `2xx`: Success (e.g., `200 OK`)
* `4xx`: Client Error (e.g., `404 Not Found`, `405 Method Not Allowed`)
* `5xx`: Server Error (e.g., `500 Internal Server Error`)

If a client requests a path that your server doesn't know about, returning a `404 Not Found` is the correct semantic response. If the path exists but the client uses an unsupported HTTP method (like sending a `POST` request to an endpoint that only expects `GET`), the correct response is `405 Method Not Allowed`.

### External Resources
* [MDN: HTTP request methods](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)
* [MDN: HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)

### Your Task
Implement basic routing. If the client requests the root path `/` with a `GET` method, respond with a `200 OK`. If they request any other path, respond with a `404 Not Found`. If they request the root path `/` but use a different HTTP method (like `POST`), respond with a `405 Method Not Allowed`. All responses must be valid HTTP/1.1 messages with appropriate headers.
