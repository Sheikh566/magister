## Dynamic Routing and Responses

Static routing is useful, but modern web servers often deal with dynamic paths. In many API designs, parts of the URL are used as variables or parameters.

For instance, in the path `/users/123`, the `123` is a dynamic segment representing the user's ID.

### The Response Body

Until now, our responses have had empty bodies. When you return data, you must include two important headers:
1. `Content-Type`: Tells the client what kind of data it's receiving (e.g., `text/plain`, `application/json`).
2. `Content-Length`: Tells the client exactly how many bytes the body contains.

Without `Content-Length`, the client won't know when the response has finished downloading (unless the connection is closed, which is inefficient).

### Adding a Body to a Raw HTTP Response

In a raw HTTP response, the body comes *after* the blank line that ends the headers. You write: the status line, the headers (each followed by `\r\n`), the blank line (`\r\n\r\n`), and then the body bytes. For example, to echo the text `hello`:

```http
HTTP/1.1 200 OK\r\n
Content-Type: text/plain\r\n
Content-Length: 5\r\n
\r\n
hello
```

The `Content-Length` must be the exact byte count of the body. For UTF-8 text, each character is typically one byte; use the length of the string in bytes (e.g., `len(body)` in Go) rather than a hardcoded value.

### External Resources
* [MDN: MIME types (Content-Type)](https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types)
* [MDN: Content-Length](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Length)

### Your Task
Create a dynamic route that matches `/echo/<text>`. When a client makes a `GET` request to this route, extract the `<text>` segment and return it as the response body. 

To ensure you aren't hardcoding a specific response, the tester will generate a random string for the path segment on every run (e.g. `GET /echo/a1b2c3d4`).

Set the `Content-Type` to `text/plain` and accurately calculate the `Content-Length`.
