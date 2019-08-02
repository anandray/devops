# Cloudstore Interface

The following interface is mapped to Ceph + LevelDB backends, with NewCloudstore setting up the connection using a config

```
type Cloudstore interface {
	GetChunk(k []byte) (v []byte, ok bool, err error)
	GetChunkBatch(chunks []*RawChunk) (err error)
	GetChunkWithRange(k []byte, start int, end int) (v []byte, ok bool, err error)
	SetChunk(ctx context.Context, k []byte, v []byte) (err error)
	SetChunkBatch(ctx context.Context, chunk []*RawChunk) (err error)
	Close()
}
```

# Tests

This tests LevelDB and Ceph with the same 4 tests (SetChunk, GetChunk, SetChunkBatch, GetChunkBatch) defined in the `Cloudstore` interface:

```
root@d3:~/go/src/github.com/wolkdb/cloudstore/wolk/cloud# go test
d3|INFO [08-01|19:28:22.645] New Cloudstore: Leveldb                  dd=/usr/local/wolk caller=cloud.go:51
d3|INFO [08-01|19:28:22.646] [cloud_test] SetChunk PASS               provider=leveldb w=47.002µs caller=cloud_test.go:69
d3|INFO [08-01|19:28:22.646] [cloud_test] GetChunk PASS               provider=leveldb r=6.101µs caller=cloud_test.go:81
d3|INFO [08-01|19:28:22.646] [cloud_test] SetChunk PASS               provider=leveldb r=71.052µs caller=cloud_test.go:91
d3|INFO [08-01|19:28:22.646] [cloud_test] GetChunk 0-byte value PASS  provider=leveldb r=3.026µs  caller=cloud_test.go:105
d3|INFO [08-01|19:28:22.646] [cloud_test] GetChunk 0-byte value PASS  provider=leveldb r=76.046µs caller=cloud_test.go:119
d3|INFO [08-01|19:28:22.647] [cloud_test] SetChunkBatch of 0-byte and N-byte value PASS provider=leveldb w=145.485µs caller=cloud_test.go:143
d3|INFO [08-01|19:28:22.647] [cloud_test] GetChunkBatch 0-byte value PASS provider=leveldb caller=cloud_test.go:169
d3|INFO [08-01|19:28:22.647] [cloud_test] GetChunkBatch NORMAL key PASS provider=leveldb caller=cloud_test.go:179
d3|INFO [08-01|19:28:22.647] [cloud_test] GetChunkBatch MISSING key PASS provider=leveldb caller=cloud_test.go:189
d3|INFO [08-01|19:28:22.648] [cloud_test] SetChunkBatch               provider=leveldb w=1.043826ms caller=cloud_test.go:209
d3|INFO [08-01|19:28:22.648] [cloud_test] GetChunkBatch               provider=leveldb w=75.869µs   caller=cloud_test.go:234
d3|INFO [08-01|19:28:22.648] New Cloudstore: ceph                     POOL=z0 caller=cloud.go:53
d3|INFO [08-01|19:28:22.673] [cloud_test] SetChunk PASS               provider=cephif  w=7.738837ms caller=cloud_test.go:69
d3|INFO [08-01|19:28:22.675] [cloud_test] GetChunk PASS               provider=cephif  r=1.648028ms caller=cloud_test.go:81
d3|INFO [08-01|19:28:22.681] [cloud_test] SetChunk PASS               provider=cephif  r=7.765254ms caller=cloud_test.go:91
d3|INFO [08-01|19:28:22.682] [cloud_test] GetChunk 0-byte value PASS  provider=cephif  r=1.330284ms caller=cloud_test.go:105
d3|INFO [08-01|19:28:22.686] [cloud_test] GetChunk 0-byte value PASS  provider=cephif  r=3.823928ms caller=cloud_test.go:119
d3|INFO [08-01|19:28:22.686] [cloud_test] SetChunkBatch of 0-byte and N-byte value PASS provider=cephif  w=3.93503ms  caller=cloud_test.go:143
d3|INFO [08-01|19:28:22.775] [cloud_test] GetChunkBatch 0-byte value PASS provider=cephif  caller=cloud_test.go:169
d3|INFO [08-01|19:28:22.775] [cloud_test] GetChunkBatch NORMAL key PASS provider=cephif  caller=cloud_test.go:179
d3|INFO [08-01|19:28:22.775] [cloud_test] GetChunkBatch MISSING key PASS provider=cephif  caller=cloud_test.go:189
d3|INFO [08-01|19:28:22.775] [cloud_test] SetChunkBatch               provider=cephif  w=29µs       caller=cloud_test.go:209
d3|INFO [08-01|19:28:22.889] [cloud_test] GetChunkBatch               provider=cephif  w=14.543619ms caller=cloud_test.go:234
PASS
ok  	github.com/wolkdb/cloudstore/wolk/cloud	0.303s

```

