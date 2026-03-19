## The HEAD Method

The `HEAD` method asks for a response identical to a `GET` request, but without the response body. This is extremely useful for clients that want to check the status or metadata of a resource before downloading it.

For example, a browser might send a `HEAD` request to a large video file. If the server returns `Content-Length: 1000000000` and `Content-Type: video/mp4`, the browser can decide whether or not to prompt the user or stream the file, without actually downloading a gigabyte of data upfront.

### Correct HEAD Implementation

When responding to a `HEAD` request, the server must calculate the headers *exactly* as it would for a `GET` request to the same path. 
Crucially, it must calculate and send the correct `Content-Length` header, but it must **not** send the actual body bytes.

### External Resources
* [MDN: HEAD Method](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/HEAD)
* [RFC 9110: HEAD](https://httpwg.org/specs/rfc9110.html#HEAD)

### Your Task
Implement support for the `HEAD` HTTP method. For any path that your server currently supports via `GET` (like `/echo/<text>`), a `HEAD` request should return the exact same status code and headers (including `Content-Length`), but omit the body entirely.
