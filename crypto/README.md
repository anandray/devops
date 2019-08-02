
```
$ go test -run Thousand
=== RUN   TestThousand
2019/03/11 19:36:43 Time required to sign 1000 messages and aggregate 1000 pub keys and signatures: 0.250360 seconds
2019/03/11 19:36:43 Aggregate Signature: 0x1 196a94ef4233ea38893c3d15e7976d2e8c42002ecdcdc39f6d367afafd77f0d5 20e12e1df9d94755d8bb0520408c5925d1b5dfb0a611f5e25dca3e221cf702b7
2019/03/11 19:36:43 Aggregate Public Key: 0x1 1c0779d391c770600de1cc161ace1362630490d989b5f38e015eebf9b1892868 7c843c1e63a2598bc17e7a0b98f3ac9c6089f7ed6dca8479fef15bf66d76a7e 195b6145a134b3476ec29544d5031ec01c082660ed4a43161c0b20ce0fb21171 1de70deb09c4db90578e34c5cdb153c56fefeda05d55a63a14b78741d9923dd8
2019/03/11 19:36:43 Aggregate Signature Verifies Correctly!
2019/03/11 19:36:43 Time required to verify aggregate sig: 0.000968 seconds
--- PASS: TestThousand (0.25s)
    bls_test.go:534: unitN=4
PASS
ok  github.com/wolkdb/cloudstore/crypto/bls0.259s
```
