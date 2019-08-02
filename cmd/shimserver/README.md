# Docker Setup

1. [Install](https://github.com/wolkdb/wolkjs#installing-wcloud) the Wolk Command Line client (wcloud)
2. Navigate to the `wolkjs` directory (`cd wolkjs`)
3. Determine your account name (`<accountName>`) 
4. Run the createaccount command to generate your public and private keys by running the following:

```./wcloud createaccount <accountName>```

5. Clone this `shimserver` repository
6. Navigate to the `shimserver` directory (`cd shimserver`)
7. Copy the private and public keys (generated in step 4) into the shimserver directory

```mkdir -p ./keys && cp ~/.wolk/rod/public.key "$_"```

```mkdir -p ./keys && cp ~/.wolk/rod/private.key "$_"```

8. Build Docker:

```
docker build -t wolk-shimserver .
```

9. Run Docker:

```
docker run --name=wolk-shimserver -p 99:99 wolk-shimserver
```

The shimserver by default uses port 99.  Replace the exposed port with the desired port for your setup if necessary.  For example, if you would like to access via port 9000 the docker command would be:

```
docker run --name=wolk-shimserver -p 9000:99 wolk-shimserver
```

10. To confirm the shimserver is up and running successfully, you can pick any website, and append it to the end of the server and exposed port combination.

For example, if the docker is running on `test.server.com` and port `99` is exposed, when you navigate to `http://test.server.com:99/espn.com`, it should show the same thing as seen when going to `http://espn.com`

# Shim URLs

Legacy "Centralized" Web servers are easy to censor, but the content of these legacy web servers can be replicated with a simple  extension.
Through this mechanism, any legacy web server can store any content on the Wolk blockchain.

To make a request to the Wolk URL like 
```
wolk://archive.org/28/items/MartinLutherKing-IHaveADream/MartinLutherKing-IHaveADream.mp3
```
return the content stored at 
```
https://gateway.archive.org/28/items/MartinLutherKing-IHaveADream/MartinLutherKing-IHaveADream.mp3
```
we simply add a `shimUrl` (in the above example https://gateway.archive.org) as a new attribute in the `TxBucket` structure, specified in account creation or updated later (via PUT/PATCH calls).
This `shimUrl` can be a http or https base URL, but could also be an S3, GC, IPFS, FTP url as well,
and is used to map wolk urls to another space.  

So, when an HTTP GET request to
```
wolk://en.xyz.com/a/b/c
```
the end user is expecting to view the same contents available at `http(s)://xyz.com/a/b/c` (i.e. the legacyUrl).

If upon making the above request the contents are not available from Wolk Cloudstore (not expired), the wolk node receiving this request follows up with a request to the `shimUrl` (for example: https://shim.xyz.com) along with the domain and path of the legacyUrl:
```
http://shim.xyz.com/en.xyz.com/a/b/c
```

The code in this repository is designed to make a request to the legacy url to retrieve this content and then sign it with the keys associated with the the account owner 'en.xyz.com'.

```
USER ==wolk://en.xyz.com/a/b/c + sig/requester GET Headers====> WOLK NODE ==on miss, requests http://shim.xyz.com/en.xyz.com/a/b/c + sig/requester GET headers==> LEGACY GATEWAY SERVER
```

The gateway server should return back the data SIGNED in the same way as an HTTP PUT would be, with the Wolk standard headers, for inclusion in the blockchain:

```
Sig: 64 character Hex string
Requester: {..JWK PublicKey of IA..}
```

After verification, the wolk node receiving the content returns the content back to the user:

```
USER <==returns content + recent tx==== WOLK NODE submits tx to store content (using sig/requester returned by LEGACY GATEWAY SERVER) <==returns content + sig/requester PUT headers for use submitting tx== LEGACY GATEWAY SERVER
```

The user's browser can verify that the transaction of the gateway server matches that of the originator.
In this way, the Wolk Node acts as an HTTP proxy, the Wolk node is requesting the content of the gateway rather than the user.

If the gateway server has Expires / Cache-Control headers in its response, the Wolk Node should record these headers in the tx.
This implies that certain headers must be included in payload bytes.  If a Legacy URL has expired, the wolk node can retrieve it again.
If the owner of a Wolk account changes their legacyUrl, this could be considered to expire all content in the Wolk account.

Gateway servers may implement the above mechanism with an Apache Wolk Module that:
 (a) reads headers [Sig, Requester] from the Wolk Node and validates them against the URL provided;
 (b) uses the private key to execute the HTTP PUT operation with the Wolk Blockchain, storing in their own bucket
The Gateway module must simply specify the JWK Private Key in their Apache Module.


## Transport

The IA Transport Class interfaces with the Wolk class by mapping urls into signed HTTP GET requests.  
* Only 200 level responses are recorded in the Wolk bucket.  
* If the document being requested is not found, a 404 is returned.
* If users have range requests, however, the entire document is requested and stored

## Development/Test

* shim(owner, collection, key) in `backend_nosql.go`
* shimserver.go implementation
* TestShim client test:
 1. Client does signed HTTP GET of `wolk://accountxyz/malala.mp4`
 2. Wolk node sends to http://gateway.plasmabolt.com/malala.mp4
 3. shimserver signs HTTP PUT and returns malala.mp4 content
 4. Content of wolk://accountxyz is available

# Business Model

Any content that has infrastructure and is censorable could be shimmed in this way.
* Archive.org
* Github, Docker
* Wikipedia
* LinkedIn, Facebook
* CNN, ...

However, there is no requirement that the gateway server be next to the filesystem holding the content.
The gateway server can simply crawl the content and deliver it signed.
