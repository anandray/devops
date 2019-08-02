# Wolk Storage / Consensus Refactorization

This quarter we are going to solve the following:
 * use Kubernetes as a node architecture and Ceph as a storage engine, which will reduce node costs and support scalability
 * connect proof of storage to rewards mechanism
 * implement storer selection using VRFs
 * factorize distributed consensus and storage

### Background
Over the last 2 years, we have been developing software for distributed storage to power blockchains for decentralized web.
The simple idea that we have reduced to practice is to have key-value pairs submitted as transactions in a blockchain state
(e.g. user123's "videos" collection like wolk://user123/videos/bday.mp4 is proven as valid as a given block with SMT proofs)
and for blocks, transactions, and the state data stored in distributed storage.
Traditionally, blockchain state has been kept purely locally, but Wolk has demonstrated it is reasonable to
connect distributed consensus to distributed storage.  The model we have been pursuing is that a block and its chunks
(chunks representing SMT+AVL trees, tx chunks, files in SetName/SetBucket/SetKey operations) are kept by a set of registered
nodes who are held responsible by the chunkIDs.  The approach of having a registry is far more efficient than the DHT type approaches because
it leads to bounded time of upload/downloads, and this is particularly critical for representing trees.

Having distributed storage backing the blockchain state must be compared to the canon of blockchain has nodes recording the entire chain from the genesis block onwards in local storage.
The ideal of users running their own nodes has certainly not become the norm, and instead users have ended up trusting sites like infura and etherscan much in the same way as the largest centralized web sites.
Wolk's model of having provable data storage implies that the vast majority of users AND nodes need only keep track of the most recent certified finalized blocks.
The implication is that it is necessary to have strong data availability of all blockchain state.

### Core problems to Solve

We need to solve the following problems:

1. currently, large blocks/files end up with a very large number of small chunks  (e.g. a 100MB file for 4K-32K chunk sizes ends up being 3K-24K),
For large registries of size 4K-64K, this implies that many more connections are made than necessary, especially
as our current "GetChunk" approach is to ask every potential registered source.
A basic solution to this is to have responsibility decided not chunk-by-chunk but by a file hash, so a 100MB file is kept by just 8 nodes
and only that many connections are required to retrieve the data.

2. we need to combine a proof of replication with storage rewards.  It turns out Pietrak (2018)'s model of proof of replication
can work for both file hashes and chunkIDs in a nearly identical fashion.  We currently have an issue where
when you do an HTTP PUT of a large file, a transaction can be finalized in a block but the file is not actually
immediately available -- it is kept in a queue for processing and only when the storage nodes finish is the data
actually available.

3. currently, we have responsibility for content decided by a small number of bits that index into the registered nodes.
For example, a file/chunk may have 8 different registered nodes that *could* store the chunk but if the storer doesn't
necessarily have proof.  This works well in high liveness situations, but once liveness goes down problems emerge:
A primary node that stops after 4-5 nodes claim to have replicated data will see problems with high probability if just
2 of the nodes actually do, and if these nodes have liveness issues with even some small probability quite typical of
storage arrays then users will rapidly find decentralized storage to be unusable if 1 out of 10^M transactions lead to
data loss.

# VRFs for Storer Selection

My recommendation is that we use VRFs to have nodes take responsibility for storage of blocks and files and
record responsibility in the blockchain itself using the proof of replication.  A primary node can broadcast
gossip a set of filehashes that require storage; storage nodes that receive these requests can self-select themselves to store the
data based on the filehash, download the data and store a replica and return a VRF proof of their
storage potential with a Merkle root.  The first N valid storage proofs (VRF + Merkle branch) received by the primary node
can be collated together.
 (a) When the primary node is looking to store a file/record in a SetKey transaction signed by a user, the transaction is
further signed by the primary node with the N Merkle Roots signed by the storers.  Only then is the transaction included in the blockchain.
This means there are provably N storage node that made a commitment to storing the content of the SetKey and can be held accountable.
 (b) When the primary node is looking to store a Block submitted for execution, it collects all the chunks into a list.
The proposer does the write of the Merkle roots of the collection in a key step of the voting process (e.g. the soft vote step),
and the verifiers verify the Merkle branches with the N storers.
In this way, there is provable data storage for all blockchain state and the user content referenced by it.

Putting the above concepts together, the flow for a user submitting a 100MB file might go like this:
 - user123 does a signed HTTP put to a primary node with the 100MB video "bday.mp4"
 - primary node calculates its filehash and gossips it.
 - some storage nodes recognize they can store the 100MB file, download the 100MB content and submit VRF proof of this along with a merkle root with actual replication
 - primary node takes N such VRF proofs and Merkle Roots and gossips a fully signed tx, now signed by the user and the primary node
 - a proposer includes this tx in its proposed block, which has a blockhash
 - some storage nodes recognize they can store the contents of the block, and retrieve all the contents of the block and submit a VRF proof of this along with a merkle root with their actual replication
 - the block packages all chunks into a NewChunkHash that has the chunks (txs, SMTs) and file hashes of the block (e.g. the 100MB file) along with their signed Merkle roots
 - in the consensus process, the storers in the blockchain in the "cert vote" phase, where verifiers verify the merkle branches as being correct prior to next voting and finalizing
Necessarily, these verifiable proofs of storage must have cryptoeconomic consequences for long term storage of the private user data and the public blockchain data.

## Tallying Rewards

Rewards.  On each block, the proposer can include a single valid reward transaction that verifies a block in the past cryptographically selected by a rewards verifier.
A rewards verifier submits this transaction, and selects the old block from the past using the seed of the block Q (say 100) blocks ago using a VRF
also decided by their stake.  Inside the old block is a "NewChunksHash" that has all chunks of the block and the contained SetKey data, all N merkle roots from both.
The rewards verifier audits every single chunk + filehash and submits a sample of valid branches along with a tally of nodes with valid branches and a claim of invalid.
When finalizing a block, the tallies of the rewards from the block must go into the rewarded nodes and not the ones that are invalid.
The verifier of this rewards transaction does not need to query any storer, only reverify the submitted branches, and then tally the successful
storage proofs into the nodes that have passed them.

## Tallying Punishment

Punishment.  A lazy or prejudiced verifier or a high latency connection should not result in massive punishment for a node that does not have an expected valid storage proof.
Rather than suffer economically visible consequences in an account balance, we lower the nodes external valuation.

# Kubernetes: Ceph/Rook

The usage of Ceph in Kubernetes supports a very different method of deployment of blockchain+storage.  The use of Kubernetes requires us to learn many new
skills, but the machinery is massively difference making to the overall costs of running blockchain nodes.  Because the cost of running 512+512 pods like this "manually" is basically the cost of running the compute instances, the
costs of the extra "storage" layer from the cloud provider vanish.  And, because

We have started to show the following is possible:
1. The deployment of Ceph (deployed by Rook) is remarkably easy https://github.com/wolkdb/hellostorage
2. Dockerizing go-algorand is possible: https://github.com/wolkdb/docker/blob/master/docker-algorand/Dockerfile

Previously, we had a single application `wolk` compiled from the cloudstore repo running on both "c" and "s" nodes, but now we can have 2 different Kubernetes orchestrated services:

1. Storage (from "storage") - receiving PUT transactions from users, responding to GET queries with provable queries, not doing consensus at all but just "listening"
2. Consensus (from "go-algorand") - receiving PUT transactions from the storage nodes, managing SMT/AVL state in StateDB with blocks containing roots, reaching consensus

So instead of 512 cores with "pushstaging" going out to a fixed array of servers, we have 512 pods running a "Storage" service and 512 pods running a "Consensus" service, on some port.
Instead of Supernodes that have scalable load balancers using the high level cloud services of Google BigTable/Datastore and Amazon Dynamo, we use the much cheaper "bare metal"
supported not by Filestore but the the open source mature "Ceph".  Ceph has 3 different levels we could interact with: block storage, filesystem, object storage.
I believe we can modify RADOS to adjust block device storage natively to work at the 4K chunk level and the file level using the RADOS go-ceph https://github.com/ceph/go-ceph library:
 - Read https://godoc.org/github.com/ceph/go-ceph/rados#IOContext.Read
 - WriteFull https://godoc.org/github.com/ceph/go-ceph/rados#IOContext.WriteFull
where the data and the proof of storage are kept by a Storage node with a Ceph persistent volume.
As a stepping stone to that, we can get started just doing what Filestore did, keeping 1 file per chunk and 1 file per file,
and then showing that in a non-Kubernetes standard way that we can interact with the Ceph volume with the above abstraction, with the first steps being:
```
# go get github.com/ceph/go-ceph
# go test -run TestRados
2019-07-05 22:57:56.438947 7f848ee18b40 -1 Errors while parsing config file!
2019-07-05 22:57:56.438954 7f848ee18b40 -1 parse_file: cannot open /etc/ceph/ceph.conf: (2) No such file or directory
2019-07-05 22:57:56.438955 7f848ee18b40 -1 parse_file: cannot open ~/.ceph/ceph.conf: (2) No such file or directory
2019-07-05 22:57:56.438955 7f848ee18b40 -1 parse_file: cannot open ceph.conf: (2) No such file or directory
unable to get monitor info from DNS SRV with service name: ceph-mon
no monitors specified to connect to.
PASS
ok  	github.com/wolkdb/hellostorage/rados	0.467s
```
Then we Dockerize the `Storage` service https://github.com/ceph/go-ceph/blob/master/Dockerfile to access the Ceph https://github.com/wolkdb/hellostorage/blob/master/manifests/stateful-rook.yaml#L39-L44

# Next Steps

Key Steps on Systems:
1. [AR,TJ] Demonstrate Kubernetes cluster can be set up across GC,AWS,Azure,Alibaba following hellostorage
2. [AR,TJ] Show that go-algorand + wolk can be deployed in Kubernetes clusters (basically Dockerizing both repositories) with any cloud provider

Key Steps on Storage:
1. [MM] Have stand alone package "storage" that can be Dockerized and put in as a Kubernetes service
2. [SN] Implement proof of storage for chunks and then files in `merkletree` package, measure performance on 100K chunks vs 100MB files
3. [SN,MM] Combine the proof of storage using "raw" file, and pass {chunk,file} tests with a Kubernetes storage service (no blockchain/consensus service)
   * GetShare/SetShare/VerifyShare
   * GetFile/SetFile/VerifyFile
4. [MM] Add RADOS Read/Write block Storage, measure performance
5. [SN,MM] Explore VRF for storer selection in Cloudstore branch where P2P messaging model and VRF model is already clear

Key Steps on Consensus:
1. [MM,SN] Connect go-algorand code to Storage Service with:
 * RegisterNode, UpdateNode working in concert with SRV bootstrap mechanism
 * Implement VRF based Storer selection
2. [AC] Fuse SMT/AVL/StateDB code into go-algorand "ledger" package
3. [MC] Ensure preemptive query machinery works
4. [MC] Connect WriteBlock + Rewards activity to Storage Proof tallies
