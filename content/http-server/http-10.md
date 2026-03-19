## Serving Static Files

One of the most fundamental jobs of a web server is to serve static files from a disk. Whether it's HTML, CSS, JavaScript, or images, the server reads bytes from a file system and streams them back to the client over HTTP.

### How it Works

When a client makes a request to a static file route (e.g., `GET /files/hello.txt`), your server must:
1. **Find the root directory:** Your server should be configured to serve files from a specific base directory. For this lesson, the tester will pass this directory to your server as an environment variable.
2. **Construct the path:** Combine the base directory with the requested filename to find the file on disk (e.g. `/tmp/www/hello.txt`).
3. **Read the file:** Open the file and read its contents.
4. **Determine the size:** Find out how many bytes the file contains so you can set the `Content-Length` header.
5. **Send the response:** Send a `200 OK` response with the `Content-Length` header, followed by an empty line, and finally the exact bytes of the file.

If the file does not exist, you must gracefully return a `404 Not Found` response.

### Directory Traversal Security

When exposing files, security is paramount. If your server blindly serves any path requested by the user, an attacker could request `GET /files/../../../etc/passwd` and read sensitive system files!

A static file server is usually configured with a **Root Directory**. It must ensure that all requested paths are strictly inside this root directory, preventing directory traversal attacks. By simply looking for `<filename>` inside your configured directory, you can safely scope the request to allowed files.

### Example Response

If the client requests a file that contains the text `Hello, World!`, your raw HTTP response should look like this:

```http
HTTP/1.1 200 OK
Content-Length: 13

Hello, World!
```
*(Remember that every line in the headers must end with `\r\n`, and there must be an empty line `\r\n\r\n` separating the headers from the body).*

### External Resources
* [MDN: Serving Static Files](https://developer.mozilla.org/en-US/docs/Learn/Server-side/Express_Nodejs/skeleton_website) (Conceptual)
* [OWASP: Path Traversal](https://owasp.org/www-community/attacks/Path_Traversal)

### Your Task
Implement a static file server. The tester will provide a directory path to your process via an environment variable (`FILES_ROOT` or `COURSE_FILES_ROOT`).

When a client requests `GET /files/<filename>`, you must look for `<filename>` inside that configured directory.
- If the file exists, read it from disk and return its contents with a `200 OK` and a matching `Content-Length`.
- If the file does not exist, return a `404 Not Found`.