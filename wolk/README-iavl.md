
# Current Node structure

Basic IAVL read/write test is passing:

```
Sourabhs-iMac:wolk sourabh$ go test -run TestKVChainIAVL
[test] iavl: bn(2) maxkeys(0) blockhash(0000000000000000000000000000000000000000000000000000000000000000)
[test] iavl: bn(3) maxkeys(1000) blockhash(bd1124bc97481217607fbf00aba2bff4d76f0611f3010573abed1d1dd4ad11be)
[test] iavl: bn(4) maxkeys(2000) blockhash(db03c82280b03ca485e27567ebedc7a235536f3208b1ad02d32a5bd2eba5914e)
[test] iavl Read Time: 112.056326ms	Reads:1000	success: 1000	notfound:0
[test] iavl: bn(5) maxkeys(3000) blockhash(cbf87c138f860826894f2d82c0e61b88e27b72adff60452b11c54709f012134c)
[test] iavl Read Time: 132.273116ms	Reads:1000	success: 2000	notfound:0
[test] iavl: bn(6) maxkeys(4000) blockhash(9aecb7218a8a1172803554b3295df6fe4a1431dd1522cba8e4b0918e4c526597)
[test] iavl Read Time: 137.323673ms	Reads:1000	success: 3000	notfound:0
...
PASS
ok  	github.com/wolkdb/cloudstore/wolk	3.926s
```
but proofs are not:
```
Sourabhs-iMac:wolk sourabh$ go test -run TestIAVLProof
--- FAIL: TestIAVLProof (0.01s)
	kvchain_test.go:249: Verify ERR: invalid root
FAIL
exit status 1
FAIL	github.com/wolkdb/cloudstore/wolk	0.053s
```

Here are notes as to what the current state is:

`SaveNode` uses `writeBytes` which has:
 * leaf nodes: writes of key+valHash
 * branch nodes: writes of key only

`MakeNode` has:
 * leaf nodes: reads key+valHash
 * branch nodes: reads  key

I used `writeBytes` instead of `writeHashBytes` which has this instead:
 * leaf nodes:  writes key+valHash
 * branch nodes: writes key+valHash
which couldn't fit with the above.

So to be consistent, I have the following:
 1. `immutable.Hash()` which uses
 2. `hashWithCount which uses
 3. `writeHashBytesRecursively` which uses `writeBytes`

BUT on Proofs:
 1. leaf nodes: writes key + valHash  [see `iavl_proof.go: func (pln proofLeafNode) Hash() []byte`]
 2. branch nodes: does NOT write the key [see `iavl_proof.go: func (pin proofInnerNode) Hash(childHash []byte) []byte`]

The current mystery is how proofs do not have keys written in the branch nodes.

# Original IAVL

`MakeNode` https://github.com/tendermint/iavl/blob/master/node.go#L67-L87
 * leaf:   key + value
 * branch: key only

`SaveNode` uses writebytes https://github.com/tendermint/iavl/blob/master/node.go#L313-L323
 * leaf:   key + value
 * branch: key only


`SaveVersion` calls `SaveBranch` https://github.com/tendermint/iavl/blob/master/mutable_tree.go#L350
`SaveBranch` calls `_hash` before calling `SaveNode` https://github.com/tendermint/iavl/blob/master/nodedb.go#L140-L148
{_hash https://github.com/tendermint/iavl/blob/master/node.go#L204, writeHashBytesRecursively} uses writeHashBytes
 * leaf:   key + valueHash
 * branch: none

Proofs:
 * `proofInnerNode`: no key https://github.com/tendermint/iavl/blob/master/proof.go#L64
 * `proofLeafNode`: key+valueHash https://github.com/tendermint/iavl/blob/master/proof.go#L123-L128


# Next Steps

1. get Proofs passing in addition to GetKey
2. Add storageBytes to node structure



# readme

AVL tree is a BST (Binary Search Tree) that "rotates" to keep balanced. The different in heights between the left and right side of a node must be [-1, 0, 1].
example: https://www.geeksforgeeks.org/avl-tree-set-1-insertion/
The method is first to insert according to BST rules (by key) then balance the tree through recursive rotations until any potential < -1 or > 1 difference in heights is resolved:

Setting:
`tree.recursiveset`
https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_mutable_tree.go#L108
sets up the tree and inserts the new node. However it does not recompute all the hashes, it just sets up (recursively) and BALANCES the tree. this is needed before computing all the hashes, because the paths may change.

The merkleization comes next, in the recursive computation of the hashes of the nodes once the structure is set:

Flushing:
`Flush` -> `SaveVersion` -> `SaveBranch` (calls `SaveNode` recursively)
flush https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_tree.go#L37
saveversion: https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_mutable_tree.go#L347
savebranch: https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_nodedb.go#L173
savenode: https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_nodedb.go#L116
SaveBranch calculates the left and right hashes of the nodes recursively, and then hashes the updated node (`node._hash()`) using writeHashBytes before saving it.  The tree is saved this way.

[optimization note: back in SaveVersion, it returns the tree.Hash() which is just the recursive hash of the root node. This is the same as the method above, so there is cleanup needed here, as they have proposed and is more obvious for us b/c we are not using two methods to hash.]

note: For each node, there is storageBytes and storageBytesTotal. The number of storage bytes we assign to a k,v pair is storageBytes. storageBytesTotal is the sum of all the storageBytes in a node's subtrees, including the node itself.
The calculation for the totals is done in recursive `SaveBranch`.

Proofs are created when a `Get` key is called:

`Get` -> proof_range:`GetWithProof` -> `getRangeProof`
get: https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_tree.go#L57
getwithproof: https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_proof_range.go#L481
getRangeProof: https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_proof_range.go#L14

A RangeProof is a leftpath, one array of inner nodes, one array of leaf nodes, and a couple of flags.
rangeproof: https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_proof_range.go#L14
right now the only difference between proof leaf nodes and proof inner nodes is that leaf nodes contain the value where inner nodes contain left and right child hashes.

It takes the existing tree and recomputes all the hashes from the rootnode, to make sure (t.root.hashWithCount).
Then we compute a PathToLeaf, using the keystart.  
https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_proof.go#L189
Then we recursively travel through the tree, using that node as a starting point, collecting inner nodes and leaf nodes along the way, also collecting keys and values within the key range specified.

Verifying Proofs:
https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_tree.go#L190
There are 2 parts to verifying proofs. The root of the proof must be verified, and a (k,v) pair must be verified.

Root verification:
Verify: https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_proof_range.go#L183
computeRootHash: https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_proof_range.go#L233
`_computeRootHash` recursively computes the hashes based on the path's inner nodes and leaf nodes, verifying intermediate hashes along the way. it returns the final root hash, which is checked against what was initially believed to be the root hash.  The verification does 0 GetNode's.

Item verification:
verifyitem: https://github.com/wolkdb/cloudstore/blob/iavlproof/wolk/iavl_proof_range.go#L95
VerifyItem searches for the key leaf in the proof path, and checks the value.
