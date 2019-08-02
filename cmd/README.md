
# Basic Storage Tests

## curl Test

### share

```
$ curl "http://c0.wolk.com/wolk/share" --data-binary @chunk
{"h":"93ca526f37c054a7a0a221f016f9346805b980f634378c7e1499b3e2e7a7a599","mr":"93ca526f37c054a7a0a221f016f9346805b980f634378c7e1499b3e2e7a7a599", "len":4096 }Sourabhs-iMac-2:ab sourabhniyogi$
$ curl "http://c0.wolk.com/wolk/share/93ca526f37c054a7a0a221f016f9346805b980f634378c7e1499b3e2e7a7a599"
<raw binary data>
```

### chunk

```
$ curl "http://c0.wolk.com/wolk/chunk/" --data-binary @chunk
{"h":"93ca526f37c054a7a0a221f016f9346805b980f634378c7e1499b3e2e7a7a599","mr":"93ca526f37c054a7a0a221f016f9346805b980f634378c7e1499b3e2e7a7a599", "len":4096 }Sourabhs-iMac-2:ab sourabhniyogi$
$ curl "http://c0.wolk.com/wolk/chunk/93ca526f37c054a7a0a221f016f9346805b980f634378c7e1499b3e2e7a7a599"
<raw binary data>
```

### batch

```
$ cat setbatch.json
[{"chunkID":"++1jAtmqZ92ocsfj2U9SZp1VOlFXUc9aWNBoSi87WNU=","val":"ipYU+Z09EkpK5/iCsKjIZDgtEmrNT7qJSSsH9wWOEME="},{"chunkID":"h8ctFvSKy2RP7tR55sncZ/SuOirFFYsbHfRtHTjVXes=","val":"aZGwJHQH8X66vIcbRXtzVJOXRpcTae+TdQNSA5pwP7I="},{"chunkID":"ymw42Bmc1VVcstJTww+N+jVpjZIz7SlcYSVvEJkgfbY=","val":"TxneQFJXUdRrfyi5wryg8ie+Z2cz2H5KHGEgJzuvVeQ="},{"chunkID":"fPPSyMqF4j+URYUwk9DXs2/c8MymvTWrGuefyYHXHME=","val":"ON3FaVeSRw/xy3zvBdU71C/dCmlbbLMsszvpNvzYrjU="},{"chunkID":"kY8DRZVjwJJ48q+kGSScij0cdRdUimuaFDY0BEH9yTw=","val":"xbhoL+DiVZVIjGLvb6OmVBhaXqZ/vY48n9PIkgEoUcE="}]
$ curl "http://c0.wolk.com/wolk/setbatch" -d @setbatch.json
{"err":"<nil>"}

$ cat getbatch.json
[{"chunkID":"++1jAtmqZ92ocsfj2U9SZp1VOlFXUc9aWNBoSi87WNU="},{"chunkID":"h8ctFvSKy2RP7tR55sncZ/SuOirFFYsbHfRtHTjVXes="},{"chunkID":"ymw42Bmc1VVcstJTww+N+jVpjZIz7SlcYSVvEJkgfbY="},{"chunkID":"fPPSyMqF4j+URYUwk9DXs2/c8MymvTWrGuefyYHXHME="},{"chunkID":"kY8DRZVjwJJ48q+kGSScij0cdRdUimuaFDY0BEH9yTw="}]
$ curl "http://c0.wolk.com/wolk/getbatch" -d @getbatch.json
[{"chunkID":"++1jAtmqZ92ocsfj2U9SZp1VOlFXUc9aWNBoSi87WNU=","val":"ipYU+Z09EkpK5/iCsKjIZDgtEmrNT7qJSSsH9wWOEME=","ok":true},{"chunkID":"h8ctFvSKy2RP7tR55sncZ/SuOirFFYsbHfRtHTjVXes=","val":"aZGwJHQH8X66vIcbRXtzVJOXRpcTae+TdQNSA5pwP7I=","ok":true},{"chunkID":"ymw42Bmc1VVcstJTww+N+jVpjZIz7SlcYSVvEJkgfbY=","val":"TxneQFJXUdRrfyi5wryg8ie+Z2cz2H5KHGEgJzuvVeQ=","ok":true},{"chunkID":"fPPSyMqF4j+URYUwk9DXs2/c8MymvTWrGuefyYHXHME=","val":"ON3FaVeSRw/xy3zvBdU71C/dCmlbbLMsszvpNvzYrjU=","ok":true},{"chunkID":"kY8DRZVjwJJ48q+kGSScij0cdRdUimuaFDY0BEH9yTw=","val":"xbhoL+DiVZVIjGLvb6OmVBhaXqZ/vY48n9PIkgEoUcE=","ok":true}]
```

## Apache Benchmark (ab) test

See `ab/ab.sh` which takes the above curl calls and does it repeatedly

## Wolk Benchmark (wb) test

The `wb` program benchmarks wolk's setshare/getshare and setchunk/getchunk interfaces by sending `n` requests in a stream of `c` concurrent go routines and shows the amount of time for:

1. `n` set calls
2. `n` get calls (no caching)
3. `n` get calls (caching)

The `run` parameter decides whether the above set/get pattern uses `share` or `chunk` and the `rpc` flag decides whether to use the RPC or http interface.

### Usage:

* Wolk Benchmark HTTP:
```
./wb -n=200 -c=10 -server c0.wolk.com -run=share
./wb -n=200 -c=10 -server c0.wolk.com -run=chunk
```

Example Trace:
```
$ ./wb -n=200 -c=10 -server 35.245.222.192 -run=share
Wolk Benchmark Server 35.245.222.192 [c=10 run=share, n=200 Chunks, chunkSize=4096 bytes]
share
Phase 0 - setshare
50%:  56
66%:  57
75%:  58
80%:  58
90%:  60
95%:  65
98%:  72
99%:  72
100%: 73
errmap: map[200:200] - elapsed: 1.170512808s Success:200

share
Phase 1 - getshare
50%:  26
66%:  27
75%:  28
80%:  28
90%:  31
95%:  50
98%:  55
99%:  55
100%: 56
errmap: map[200:200] - elapsed: 0.594866057s Success:200

share
Phase 2 - GET share
50%:  26
66%:  27
75%:  27
80%:  28
90%:  28
95%:  50
98%:  55
99%:  55
100%: 55
errmap: map[200:200] - elapsed: 0.588942356s Success:200


$ ./wb -n=200 -c=10 -server cloudstore-us-central-gc.wolk.com -run=chunk
Wolk Benchmark Server cloudstore-us-central-gc.wolk.com [c=10 run=chunk, n=200 Chunks, chunkSize=4096 bytes]
chunk
Phase 0 - POST chunk
50%:  402
66%:  478
75%:  571
80%:  600
90%:  760
95%:  818
98%:  886
99%:  942
100%: 1005
errmap: map[200:200] - elapsed: 9.082178477s Success:200

chunk
Phase 1 - GET chunk
50%:  174
66%:  188
75%:  206
80%:  221
90%:  255
95%:  342
98%:  625
99%:  673
100%: 759
errmap: map[200:200] - elapsed: 3.651727887s Success:200

chunk
Phase 2 - GET chunk
50%:  149
66%:  194
75%:  217
80%:  235
90%:  264
95%:  296
98%:  354
99%:  387
100%: 426
errmap: map[200:200] - elapsed: 3.090672071s Success:200
```
