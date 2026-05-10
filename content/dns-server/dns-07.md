## Handling Name Compression

DNS packets often reuse the same domain names (for example, in the question section and then again in the answer section). To save bytes, DNS uses a compression scheme where a domain name (or a part of it) can be replaced with a two-byte pointer to an earlier occurrence of that name in the same message.

### Identifying a Pointer

Recall that a standard domain name is a sequence of labels, where each label starts with a 1-byte length. 

Because a single label is limited to 63 characters, the length byte will never be larger than `63` (which is `00111111` in binary). DNS takes advantage of this by using the top two bits of the length byte to indicate whether it's a standard length or a pointer:

* If the top two bits are `00`, it's a standard label length.
* If the top two bits are `11` (or `0xC0` in hex), it's the start of a 2-byte compression pointer.

### Pointer Layout

A compression pointer is exactly 2 bytes (16 bits) long. The first two bits are `11`, and the remaining 14 bits represent the offset from the very beginning of the DNS packet (where the ID field starts).

```text
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
| 1  1|                OFFSET                   |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

For example, if you encounter the bytes `C0 0C`:
1. In binary, `C0 0C` is `11000000 00001100`.
2. The top two bits are `11`, confirming it's a pointer.
3. Masking out the top two bits leaves `00000000 00001100`, which is `12` in decimal.
4. This means the actual domain name starts at byte offset `12` of the DNS packet (which happens to be immediately after the 12-byte header).

### How to Parse Compressed Names

When reading a domain name, you should check the first byte of each label:
1. If it's a normal length (top bits `00`), read the label and continue.
2. If it's a pointer (top bits `11`), read the second byte, calculate the 14-bit offset, and jump to that offset in the packet to continue reading the name.
3. A pointer always marks the end of the current domain name string. You do not need to look for a null byte after a pointer. Once you resolve the pointer and read the rest of the name from the new location, you are done with this domain name.

### Safety Notes

Compression pointers can be malformed. Your decoder should reject pointers that leave the packet or loop forever. A simple way to prevent infinite loops is to limit the maximum number of pointer jumps allowed per name (e.g., 5 or 10 jumps).

### External Resources

* [RFC 1035: Message compression](https://www.rfc-editor.org/rfc/rfc1035#section-4.1.4)

### Your Task

Support compressed names in incoming questions. The tester sends two questions for `magister.test`; the second question uses a pointer to the first name. Parse both questions and return an `A` answer for each.
