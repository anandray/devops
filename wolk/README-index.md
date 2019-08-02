Having started along the sql ablation process, at the core, I'm unhappy with unprovable B+Trees (which could be replaced with IAVL Trees) but really  unprovable SQL execution (which might be solved with STARKs) and then the lack of relational support -- its 800 pounds of stuff to lift and our bench press capability is like 250 pounds.   But your support from our conversation on Sunday gives me some confidence that we should pursue the opposite direction, of adding IAVL Indexes to `TxBucket`.  My plan is to go back to the `kvtree` branch https://github.com/wolkdb/cloudstore/tree/kvtree and take this approach seriously:
1. Modify `wolk.js`'s' `createCollection` https://github.com/wolkdb/wolkjs/blob/master/apps/wolk.js#L189
```
createCollection(owner, collection, hdrs, callback)
```
to have schema+indexes from schema.org,
where hdrs would specify
(a) `Schema` from Schema.org https://schema.org/Person
(b) `Indexes`: an array of attributes from (a)
e.g.
```
Wolk.createCollection("fawkes", "friends", {"Schema":"Person", "Indexes": ["name", "email"]})
```
The default plan is to adapt SQLTable's columns into `Bucket`, to have an array of IAVLTree roothashes.
These roothashes will be initially empty.

https://github.com/wolkdb/cloudstore/blob/master/wolk/txpayload.go#L67  
https://github.com/wolkdb/cloudstore/blob/master/wolk/request.go#L29

2. Replace KVTree in StateDB [started in `kvtree` branch], modify `SetKey` to update `IAVLTree` to update `Bucket` roothashes at the end of `CommitNoSQL`.  We could have "strict" parameters added to (1) that require strong consistency to the schema of (a)


3. Modify wolk.js `scanCollection`
```
scanCollection(owner, collection, hdrs, callback)
```
to have headers of:
(a) `Range-Field`, e.g. "email"
(b) `Range-Start` and `Range-End`: e.g. "alina@wolk.com", "bruce@wolk.com"
(c) `Range-Order`, e.g. "Ascending"
(d) `Limit`
eg
```
Wolk.scanCollection("fawkes", "friends", {"Range-Field":"email", "Range-Start": "alina@wolk.com", "Range-End": "bruce@wolk.com", "Range-Order": "Ascending", "Limit": 100})
```
that would map directly into the GetRange of https://github.com/wolkdb/cloudstore/blob/9b8f8ebb23d0794220107c1f3012dccca764ea22/wolk/iavl_tree.go#L126
```func (tree *IAVLTree) GetRange(startkey []byte, endkey []byte, limit int) (keys [][]byte, values []common.Hash, proof *RangeProof, err error) {
```
with scan proofs that we can be proud of and confidence that this is a 200 pound problem.

4. The KeyBench vals https://github.com/wolkdb/cloudstore/blob/master/cmd/wolkbench/wolkbench.go#L33 in the `wolkbench` would be changed to have a lot of JSON docs conforming to standard schemas.

5. Assuming the above develops sanely, we can decide where schemas are stored, who "owns" them, how are they referenced exactly.

6. Design: DELETE / PATCH treatment

7. Design treatment of "in" on non-primary cases, general JS pattern (IndexedDB, MongoDB)

8. Develop "negative" cases: (a) invalid sort Direction (b) invalid index (c) invalid scheme
