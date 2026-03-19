## Handling Request Bodies

Just as a server can send a body in an HTTP Response, a client can send a body in an HTTP Request. This is typical for methods like `POST` and `PUT` where the client is uploading data to the server.

Here is an example `POST` request with a body:

```http
POST /mirror HTTP/1.1\r\n
Host: localhost\r\n
Content-Length: 12\r\n
\r\n
Hello, world
```

The body (`Hello, world`) appears after the blank line. The `Content-Length` header tells the server to read exactly 12 bytes.

### Knowing When to Stop Reading

When parsing an HTTP Request, you read headers until you encounter the empty line (`\r\n\r\n`). If the client intends to send a body, it will include a `Content-Length` header in the request.

Your server must:
1. Parse the headers.
2. Find the `Content-Length` value.
3. Read exactly that many bytes from the TCP connection *after* the empty line.

If you attempt to read until the connection closes, your server will hang, because modern clients keep connections open for reuse!

### External Resources
* [MDN: POST Method](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/POST)
* [MDN: Content-Length](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Length)

### Your Task
Implement a `POST /mirror` route. When a client sends a request to this route, read the request body based on the provided `Content-Length`. Then, create a `200 OK` response where the response body is an exact copy of the bytes the client sent.
