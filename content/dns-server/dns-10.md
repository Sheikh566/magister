## Caching Answers With TTL

DNS answers include a Time To Live (`TTL`), which is an integer representing the number of seconds the record is valid. To reduce network traffic and improve performance, a resolver may cache the answer while the TTL is fresh. However, it **must** stop using the cached copy after the TTL expires.

### How to Implement Caching

To implement a cache, you need a data structure (like a hash map or dictionary) that stores DNS records. 

#### 1. The Cache Key
Caching must be keyed by the specific question being asked. At a minimum, your cache key should consist of:
* **Name** (e.g., `cache.magister.test`)
* **Type** (e.g., `A` or `1`)
* **Class** (e.g., `IN` or `1`)

A cached `A` record for one name should not answer a query for a different name, nor should it answer an `AAAA` (IPv6) query for the same name.

#### 2. The Cache Value
The value stored in your cache needs to include the actual resource record data (e.g., the IP address) **and** the time it expires. 

When you receive a response from an upstream server, calculate the expiration time:
`Expiration Time = Current Time + TTL`

#### 3. Cache Retrieval
When a new query comes in:
1. Check if the `(Name, Type, Class)` exists in your cache.
2. If it does, check if `Current Time < Expiration Time`.
3. If the record is still fresh, construct a response using the cached data and send it to the client (you do not need to query the upstream server).
4. If the record has expired (or doesn't exist), forward the query to the upstream server, cache the new response, and return it to the client.

*Note: When returning a cached record, it is best practice to update the TTL in the response to reflect the remaining time (`Expiration Time - Current Time`), but for this lesson, returning the original TTL is usually sufficient.*

### External Resources

* [Cloudflare Learning Center: DNS TTL](https://www.cloudflare.com/learning/cdn/glossary/time-to-live-ttl/)
* [RFC 1035: Resource records](https://www.rfc-editor.org/rfc/rfc1035#section-3.2.1)

### Your Task

Forward queries for `cache.magister.test` to the configured upstream resolver and cache the successful answer. 

1. The upstream will return `192.0.2.55` with a TTL of `1` second. 
2. A second immediate query for the same name should be answered directly from your cache without contacting the upstream server.
3. After the 1-second TTL expires, a third query should cause your server to contact the upstream server again.
