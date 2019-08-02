# Wolk Cloudstore

Wolk Cloudstore is a Layer 1 decentralized storage blockchain where nodes are compensated for storage and bandwidth by users who use the blockchain for File Storage and NoSQL

Key Features:
* HTTP Interfaces for all operations, with Go and Javascript libraries
* Permissionless Consensus -- based on Algorand
* Low-Latency Chunk Storage

In the Solid model of the decentralized web, users control access to their data through a Solid Pod referenced by a WebID like [https://sourabhniyogi.inrupt.net/profile/card#me](https://sourabhniyogi.inrupt.net/profile/card#me).  Users use this WebID to login to Solid apps which access that data.  What we can do is to transform `cloudstore` and the Wolk blockchain protocol into a [Solid Pod provider](https://github.com/solid/solid-idp-list/blob/gh-pages/services.json) using the [Go Solid Server Code base](https://github.com/linkeddata/gold), which conforms to:
  1. [Solid Spec](https://github.com/solid/solid-spec)
  2. [Solid REST API](https://github.com/solid/solid-spec/blob/master/api-rest.md)
  3. [Linked Data Protocol](https://www.w3.org/TR/ldp/)
  4. [Solid Web Access Control](https://github.com/solid/web-access-control-spec)
and coordinates user data (RDF and non-RDF) with:
 * [Web Identity and Discovery](https://www.w3.org/2005/Incubator/webid/spec/identity/) - User profiles
 * [ACLs](https://github.com/solid/web-access-control-spec) - Authorizing WebIDs with Read/Write/Control
 * [Metadata](https://github.com/solid/solid-spec/blob/master/content-representation.md)
with applications using [WebID-OIDC](https://github.com/solid/webid-oidc-spec).
The expectation is to be able to pass the [Solid test suite](https://github.com/solid/test-suite/tree/master) that runs through [LDP tests](https://github.com/w3c/ldp-testsuite/blob/6b8da35c66f6d5f280a7507887444812cee4ae8b/src/main/java/org/w3/ldp/testsuite/test/CommonContainerTest.java#L136-L153) and [RDF Fixtures](https://github.com/kjetilk/p5-web-solid-test-basic).


### Transactions and Proofs

After [creating a user account](https://github.com/solid/solid-spec/blob/master/recommendations-client.md) with a [registered DID](https://w3c-ccg.github.io/did-spec/#public-keys), a user's pod will have several [default containers](https://github.com/solid/solid-spec/blob/master/recommendations-server.md)
* `/` private
* `/profile/` read public
* `/settings/` private, protected
* `/inbox`  append-only by public through [Linked Data Notifications](https://www.w3.org/TR/ldn/), read by owner

and can add more containers and documents to them.  Wolk blockchain transaction methods are in the `gold` repo: [CreateBucket](https://github.com/wolkdb/gold/blob/master/server.go#L938), [SetKey](https://github.com/wolkdb/gold/blob/master/server.go#L1154), [DeleteKey](https://github.com/wolkdb/gold/blob/master/server.go#L1191) and [GetKey/ScanCollection](https://github.com/wolkdb/gold/blob/master/server.go#L427-L786).   [SPARQL](https://github.com/wolkdb/gold/blob/master/server.go#L830-L832) is another mechanism.  

Key transformations for Wolk cloudstore:
 * User accounts are connected to [DID](https://w3c-ccg.github.io/did-primer/) and [WebID](https://www.w3.org/2005/Incubator/webid/spec/identity/) profiles, recorded in Blockchain
 * Every transaction should be a [signed](https://github.com/digitalbazaar/jsonld-signatures/blob/master/README.md) [JSON-LD document](https://w3c-dvcg.github.io/ld-signatures/#signature-algorithm) and then GetName, GetBuckets, GetKey, ScanCollection should return [linked data proofs](https://w3c-dvcg.github.io/ld-proofs/) mapped to SMT/AVL Proofs

### [Applications](https://github.com/solid/solid-apps)

Wolk can build a Appstore [Solid app](https://github.com/kjetilk/p5-web-solid-test-basic) following [Plume blog app](https://github.com/theWebalyst/solid-plume) from [@theWebalyst](https://forum.solidproject.org/u/happybeing/summary)) using [Solid panes](https://github.com/solid/solid-panes) and [LDflex](https://github.com/solid/query-ldflex).  Templates following React and Angular.


### More Info

Additional Resources about Solid:
 * [Inrupt](https://inrupt.net)
 * [Solid on Gitter](https://gitter.im/solid/chat?at=5be2f4407a36913a9a064514)
 * [Solid Forum](https://forum.solidproject.org/)
 * [SPARQL](https://www.youtube.com/watch?v=LUF7plExdv8)
 * [SAFE and Solid](https://safenetforum.org/t/devcon-talk-supercharging-the-safe-network-with-project-solid/23081)



# Build Wolk Miner

You can build wolk a wolk miner with: (Replace `~` with your home directory)
```
$ mkdir -p ~/src/github.com/wolkdb
$ cd ~/src/github.com/wolkdb
$ git clone git@github.com:wolkdb/cloudstore.git
$ cd cloudstore
$ export GOPATH=~
$ make wolk
build/env.sh go run build/ci.go install ./cmd/wolk
>>> /usr/local/go/bin/go install -ldflags -X main.gitCommit=fdc5f7b81073c35f1fd36efe4f6ca1c0f2f4f0a6 -s -v ./cmd/wolk
Done building.
Run "./build/bin/wolk" to launch wolk.
$ /root/go/src/github.com/wolkdb/cloudstore/build/bin/wolk --rpc --rpcaddr 0.0.0.0 --rpcapi cloudstore --port 30300 --rpcport 9900 --rpccorsdomain=* --rpcvhosts=* --wolkidx 0 --datadir /tmp/wolk
```
