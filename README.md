# WebTransport Echo Server

A tiny WebTransport echo server.

## Browser Example

```js
// get the server certificate hash
const res = await fetch('http://127.0.0.1:8000');
const hash = await res.text();
// connect using the self-signed certificate hash
const transport = new WebTransport('https://127.0.0.1:4443', {
  serverCertificateHashes: [
    { algorithm: 'sha-256', value: Uint8Array.fromHex(hash) }
  ]
});
```
