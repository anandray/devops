package wolk

import (
	//"bytes"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/pprof"

	//"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
	cloud "github.com/wolkdb/cloudstore/wolk/cloud"

	//"golang.org/x/net/netutil"
	//	opentracing "github.com/opentracing/opentracing-go"
	//	"github.com/opentracing/opentracing-go/ext"
	//	otlog "github.com/opentracing/opentracing-go/log"
	//	"github.com/wolkdb/cloudstore/wolk/log/opentrace"
	//"github.com/wolkdb/cloudstore/client/chunker"

	mlog "log"
)

const (
	httpPort    = 9800
	fullTracing = false
	useReplica  = false
)

const (
	errIncorrectLength       = "Incorrect chunkID length"
	errUserNotFound          = "User not found"
	errChunkNotFound         = "Chunk not found"
	errBucketNotFound        = "Bucket not found"
	errKeyNotFound           = "Key not found"
	errKeyDeleted            = "Key deleted"
	errInvalidFileRequest    = "Invalid file request"
	errInvalidTxRequest      = "Invalid tx request"
	errTxNotFound            = "Tx not found"
	errBlockNotFound         = "Block not found"
	errNodeNotFound          = "Node not found"
	errInvalidAccountRequest = "Invalid account request"
	errInvalidNameRequest    = "Invalid name request"
	errVerificationFailure   = "Verification failure"
	errNotImplemented        = "Not implemented"
	errTooBusyGoRoutine      = "Too Busy [Go Routine]"
	errTooBusyGeneral        = "Too Busy [General]"
	errTooBusyConnection     = "Too Busy [Connection Limit]"
	errInvalidSigner         = "Invalid signer"
)

const AccessControlExposeHeaders = "Proof,Proof-Type,Content-Length"

type HttpServer struct {
	Handler     http.Handler
	wcs         *WolkStore
	HTTPPort    int
	connections chan struct{}
	maxConns    int
	NumCPU      int
}

func NewHttpServer(wcs *WolkStore, p int, maxConns int) *HttpServer {
	if p == 80 {
		p = 443
	}
	s := &HttpServer{
		wcs:         wcs,
		HTTPPort:    p,
		NumCPU:      runtime.NumCPU(),
		maxConns:    maxConns,
		connections: make(chan struct{}, maxConns)}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.getConnection)

	// TODO: merge in cors? https://github.com/rs/cors https://github.com/wolkdb/plasma/blob/master/vendor/github.com/ethereum/go-ethereum/rpc/http.go#L221-L234
	// handler := newCorsHandler(srv, cors)
	s.Handler = mux
	for i := 0; i < maxConns; i++ {
		s.connections <- struct{}{}
	}
	log.Trace("[http:NewHttpServer]", "maxConns", maxConns, "len", len(s.connections))
	return s
}

func (s *HttpServer) Start() {
	cfg := &tls.Config{}

	cert, err := tls.LoadX509KeyPair(s.wcs.config.SSLCertFile, s.wcs.config.SSLKeyFile)
	if err == nil {
		cfg.Certificates = append(cfg.Certificates, cert)
	} else {
		log.Error("tls.LoadX509KeyPair - default", "err", err)
	}
	certFile2 := "/etc/ssl/certs/wildcard.wolk.com/www.plasmabolt.com.crt"
	keyFile2 := "/etc/ssl/certs/wildcard.wolk.com/www.plasmabolt.com.key"
	cert2, err := tls.LoadX509KeyPair(certFile2, keyFile2)
	if err == nil {
		cfg.Certificates = append(cfg.Certificates, cert2)
	} else {
		log.Error("tls.LoadX509KeyPair2", "err", err)
	}
	cfg.BuildNameToCertificate()

	go func() {
		srv := &http.Server{
			Addr:         fmt.Sprintf(":%d", s.HTTPPort),
			Handler:      s.Handler,
			ReadTimeout:  600 * time.Second,
			WriteTimeout: 600 * time.Second,
			TLSConfig:    cfg,
		}
		if len(s.wcs.config.SSLCertFile) > 0 && len(s.wcs.config.SSLKeyFile) > 0 {
			err := srv.ListenAndServeTLS("", "") // s.wcs.config.SSLCertFile, s.wcs.config.SSLKeyFile
			if err != nil {
				log.Error("[http:Start] ListenAndServeTLS error", "err", err)
			}
			log.Info("[http:Start] ListenAndServeTLS", "port", s.HTTPPort)
		} else {
			srv.Addr = fmt.Sprintf(":%d", s.HTTPPort)
			err := srv.ListenAndServe()
			if err != nil {
				log.Error("[http:Start] ListenAndServe error", "err", err)
			}
			log.Info("[http:Start] ListenAndServe", "port", s.HTTPPort)
		}
	}()

	go func() {
		var pos int
		if s.HTTPPort == 443 {
			pos = 6060
		} else {
			pos = s.HTTPPort - 80 + 6060
		}
		paddr := fmt.Sprintf("localhost:%d", pos)
		mlog.Println(http.ListenAndServe(paddr, nil))
	}()

	go func() {
		currentPort := s.HTTPPort + 6000
		currentServer := fmt.Sprintf("localhost:%d", currentPort)
		fmt.Println(http.ListenAndServe(currentServer, nil))
	}()
}

var m_count sync.RWMutex
var handlercount int64

func increaseCount() {
	m_count.Lock()
	handlercount++

	m_count.Unlock()
}

func decreaseCount() {
	m_count.Lock()
	handlercount--

	m_count.Unlock()
}

func (s *HttpServer) releaseConnection() {
	s.connections <- struct{}{}
	decreaseCount()
}

func handlerCount() int64 {
	m_count.Lock()
	defer m_count.Unlock()
	return handlercount
}

func (s *HttpServer) getConnection(w http.ResponseWriter, r *http.Request) {
	select {
	case <-s.connections:
		increaseCount()
		defer s.releaseConnection()
		s.handler(w, r)
	default:
		http.Error(w, errTooBusyConnection, http.StatusServiceUnavailable)
		log.Error("[http:getConnection] 503 Service Unavailable (getConnection)", "goroutine", runtime.NumGoroutine(), "len(s.connections)", len(s.connections), "s.maxConns", s.maxConns, "handlercount", handlerCount())
	}
}

/* TODO: one flaw we may have with the shim server stuff is that we expect the path that is used for building the message is `/owner/collection/key`
but when we send a request to a shim, it may be the case that we have a request like: wolk://dweb.archive.org/collection/key that has a
shimURL of https://dwebshim.org/ that maps to https://dwebshim.org/collection/key ... the shim server would never know or receive the
owner */

func (s *HttpServer) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Sig, Msg, Requester, Proof, SHIM, WaitForTx")
	rand.Seed(time.Now().UnixNano())

	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if strings.HasPrefix(r.URL.Path, "/pprof") {
		if len(pathpieces) < 2 {
			return
		}
		collection := pathpieces[1]
		debug := 2 // // The debug parameter enables additional output. values 0,1,2
		if len(pathpieces) >= 3 {
			debug, _ = strconv.Atoi(pathpieces[2])
		}
		switch collection {
		case "goroutine":
			s.pprotGoroutine(w, r, debug)
			break
		case "heap":
			s.pprotHeap(w, r, debug)
			break
		case "allocs":
			s.pprotAllocs(w, r, debug)
			break
		case "threadcreate":
			s.pprotThreadcreate(w, r, debug)
			break
		case "block":
			s.pprotBlock(w, r, debug)
			break
		case "mutex":
			s.pprotMutex(w, r, debug)
			break
		default:
			s.usageHandler(w, r)
		}
		return
	} else if strings.HasPrefix(r.URL.Path, "/wolk/info") {
		s.getInfoHandler(w, r)
		return
	}
	if runtime.NumGoroutine() > s.NumCPU*4000 {
		if rand.Intn(10) > 5 {
			http.Error(w, errTooBusyGoRoutine, http.StatusServiceUnavailable)
			log.Error("[http:handler] 503 error", "NumGoroutine", runtime.NumGoroutine())
			return
		}
	}
	if len(pathpieces) > 0 && (pathpieces[0] == ProtocolName) {
		if len(pathpieces) < 2 {
			return
		}
		collection := pathpieces[1]
		if r.Method == http.MethodPost {
			switch collection {
			case PayloadShare:
				s.setShareHandler(w, r)
			case PayloadChunk:
				s.setChunkHandler(w, r)
			case PayloadShareBatch:
				s.setShareBatchHandler(w, r)
			case PayloadBatch:
				s.setBatchHandler(w, r)
			case PayloadNode:
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Error("[http:submitTxHandler]", "err", err)
				}
				log.Info("[http:handler]", "collection", r.URL.Path)
				s.submitTxHandler(w, r, PayloadNode, body)
			case PayloadTransfer:
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Error("[http:submitTxHandler]", "err", err)
				}
				log.Info("[http:handler]", "collection", r.URL.Path)
				s.submitTxHandler(w, r, PayloadTransfer, body)
			}
		} else if r.Method == http.MethodGet {
			switch collection {
			case PayloadGenesis:
				s.getGenesisHandler(w, r)
			case PayloadChunkSearch:
				s.getChunkSearchHandler(w, r)
			case PayloadShare:
				s.getShareHandler(w, r)
			case PayloadChunk:
				s.getChunkHandler(w, r)
			case PayloadVote:
				s.getVoteHandler(w, r)
			case PayloadBandwidth:
				s.bandwidthHandler(w, r)
			case PayloadBatch:
				s.getBatchHandler(w, r)
			case PayloadAccount:
				s.getAccountHandler(w, r)
			case PayloadTx:
				s.getTxHandler(w, r)
			case PayloadName:
				s.getNameHandler(w, r)
			case PayloadBlock:
				s.getBlockHandler(w, r)
			case PayloadNode:
				s.getNodeHandler(w, r)
			default:
				s.usageHandler(w, r)
			}
		}
	} else if strings.HasPrefix(r.URL.Path, "/healthcheck") {
		if runtime.NumGoroutine() > s.NumCPU*3000 || s.wcs.NumPeers() == 0 {
			http.Error(w, errTooBusyGeneral, http.StatusServiceUnavailable)
			log.Error("[http:handler] NumGoRoutine check", "NumGoroutine", runtime.NumGoroutine())
		} else {
			var err error
			res := make(map[string]interface{})
			res["numpeers"] = s.wcs.NumPeers()
			res["numgoroutines"] = runtime.NumGoroutine()
			res["lastestblocknumber"], err = s.wcs.LatestBlockNumber()
			if err != nil {
				res["lastestblocknumber"] = "N/A"
			}
			res["lastFinalizedBN"] = s.wcs.LastFinalized()
			res["status"] = "OK"
			resjson, err := json.Marshal(res)
			if err != nil {
				fmt.Fprintf(w, "OK")
			} else {
				fmt.Fprintf(w, fmt.Sprintf("%s", resjson))
			}
		}
	} else if strings.HasPrefix(r.URL.Path, "/favicon.ico") {
		//?
	} else {
		options := s.getRequestOptions(r)
		owner, collection, key := s.ParseHttpUrlPath(r.URL.Path)
		//log.Info("[http:handler] url parsed to:", "owner", owner, "collection", collection, "key", key)
		switch r.Method {
		case http.MethodPatch:
			//log.Info("[http:handler] PATCH - but entering POST", "o", owner, "coll", collection, "key", key)
			s.HTTPPostMethodHandler(w, r, owner, collection, key, options)
			break
		case http.MethodPut:
			//log.Info("[http:handler] PUT - but entering POST", "o", owner, "coll", collection, "key", key)
			s.HTTPPostMethodHandler(w, r, owner, collection, key, options)
			break
		case http.MethodPost:
			//log.Info("[http:handler] POST", "o", owner, "coll", collection, "key", key)
			s.HTTPPostMethodHandler(w, r, owner, collection, key, options)
			break
		case http.MethodDelete:
			s.HTTPDeleteMethodHandler(w, r, owner, collection, key, options)
			break
		case http.MethodGet:
			s.HTTPGetMethodHandler(w, r, owner, collection, key, options)
			break
		default:
			// s.HTTPGetMethodHandler(w, r, owner, bucket, key, options)
		}
	}
}

//	goroutine    - stack traces of all current goroutines
func (s *HttpServer) pprotGoroutine(w http.ResponseWriter, r *http.Request, debug int) {
	output := "goroutine    - stack traces of all current goroutines\n"
	w.Write([]byte(output))
	pprof.Lookup("goroutine").WriteTo(w, debug)
}

//	heap         - a sampling of memory allocations of live objects
func (s *HttpServer) pprotHeap(w http.ResponseWriter, r *http.Request, debug int) {
	output := "heap         - a sampling of memory allocations of live objects\n"
	w.Write([]byte(output))
	pprof.Lookup("heap").WriteTo(w, debug)
}

//	allocs       - a sampling of all past memory allocations
func (s *HttpServer) pprotAllocs(w http.ResponseWriter, r *http.Request, debug int) {
	output := "allocs       - a sampling of all past memory allocations\n"
	w.Write([]byte(output))
	pprof.Lookup("allocs").WriteTo(w, debug)
}

//	threadcreate - stack traces that led to the creation of new OS threads
func (s *HttpServer) pprotThreadcreate(w http.ResponseWriter, r *http.Request, debug int) {
	output := "threadcreate - stack traces that led to the creation of new OS threads\n"
	w.Write([]byte(output))
	pprof.Lookup("threadcreate").WriteTo(w, debug)
}

//	block        - stack traces that led to blocking on synchronization primitives
func (s *HttpServer) pprotBlock(w http.ResponseWriter, r *http.Request, debug int) {
	output := "block        - stack traces that led to blocking on synchronization primitives\n"
	w.Write([]byte(output))
	pprof.Lookup("block").WriteTo(w, debug)
}

//	mutex        - stack traces of holders of contended mutexes
func (s *HttpServer) pprotMutex(w http.ResponseWriter, r *http.Request, debug int) {
	output := "mutex        - stack traces of holders of contended mutexes\n"
	w.Write([]byte(output))
	pprof.Lookup("mutex").WriteTo(w, debug)
}

func (s *HttpServer) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (s *HttpServer) HTTPPostMethodHandler(w http.ResponseWriter, r *http.Request, owner string, collection string, key string, options *RequestOptions) {
	// Don't allow posts to wolk://?

	if len(owner) == 0 {
		log.Error("[http:HTTPPostMethodHandler] no owner")
		http.Error(w, errNotImplemented, http.StatusNotImplemented)
		return
	}
	log.Info("[http:HTTPPostMethodHandler] ************** ", "owner", owner, "collection", collection, "key", key)

	// Why don't we check the existing owner here?

	// a new bucket is needed:
	// POST wolk://owner/ (with bucket in the body) or POST wolk://owner/collection
	if len(collection) == 0 || len(key) == 0 {
		//log.Info("[http:HTTPPostMethodHandler] creating bucket")
		s.createBucketHandler(w, r, owner)
		return
	}

	// SQL Case [PATCH=ReadSQL/POST=MutateSQL]
	// wolk://owner/database/SQL
	if key == "SQL" {
		log.Info("[http:HTTPPostMethodHandler] starting sqlhandler", "owner", owner, "bucketName", collection)
		s.sqlHandler(w, r, owner, collection, options)
		return
	}

	// bucket exists - this must be a key set
	// PUT to wolk://owner/collection/key
	bucket, ok, _, err := s.GetBucketObject(owner, collection, options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, errBucketNotFound, http.StatusNotFound)
		return
	}
	//log.Info("[http:HTTPPostMethodHandler] got existing BUCKET", "owner", owner, "collection", collection, "bucket", fmt.Sprintf("%+v", bucket))

	if len(bucket.Indexes) > 0 {
		//log.Info("[http:HTTPPostMethodHandler] bucket Indexes exist. going into setIndexedKeyHandler")
		s.setIndexedKeyHandler(w, r, bucket)
	} else {
		//log.Info("[http:HTTPPostMethodHandler] bucket Indexes don't exist. going into setKeyHandler")
		s.setKeyHandler(w, r, owner, collection, key, bucket)
	}

	return
}

/*
func (s *HttpServer) HTTPPatchMethodHandler(w http.ResponseWriter, r *http.Request, owner string, collection string, key string, options *RequestOptions) {
	// Don't allow posts to wolk://?
	if len(owner) == 0 {
		log.Error("[http:HTTPPatchMethodHandler] no owner")
		http.Error(w, errNotImplemented, http.StatusNotImplemented)
		return
	}

	// POST wolk://owner/ or POST wolk://owner/collection
	log.Info("[http:HTTPPatchMethodHandler] ************** ", "owner", owner, "collection", collection, "key", key)
	if len(collection) == 0 || len(key) == 0 { //patching an (1) account or (2) bucket
		log.Info("[http:HTTPPatchMethodHandler] ***************** creating bucket")
		s.patchBucketHandler(w, r, owner)
		return
	}

	// SQL Case [PATCH=ReadSQL/POST=MutateSQL]
	// Does it make sense to run patch SQL?
	if key == "SQL" {
		log.Info("[http:HTTPPatchMethodHandler] starting sqlhandler", "owner", owner, "bucketName", collection)
		s.sqlHandler(w, r, owner, collection, options)
		return
	}

	// POST wolk://owner/collection/key or wolk://owner/database/SQL -- get the Bucket.
	bucket, ok, _, err := s.GetBucketObject(owner, collection, options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, errBucketNotFound, http.StatusNotFound)
		return
	}

	// bucket exists
	log.Info("[http:HTTPPatchMethodHandler] BUCKET", "owner", owner, "collection", collection, "bucketName", bucket.Name, "bucketType", bucket.BucketType)

	// PUT to wolk://owner/collection/key
	// Does it make sense to run patch a key?
	s.setKeyHandler(w, r, owner, collection, key, bucket)
	return
}
*/

func (s *HttpServer) HTTPDeleteMethodHandler(w http.ResponseWriter, r *http.Request, owner string, collection string, key string, options *RequestOptions) {
	// Don't allow DELETES to wolk://
	if len(owner) == 0 {
		http.Error(w, errNotImplemented, http.StatusNotImplemented)
		return
	}

	// DELETE wolk://owner/
	if len(collection) == 0 {
		http.Error(w, errNotImplemented, http.StatusNotImplemented)
		//could be a case for using "deleteName"?
		return
	}

	// DELETE wolk://owner/collection/key
	s.deleteKeyHandler(w, r, owner, collection, key)
	return
}

func (s *HttpServer) HTTPGetMethodHandler(w http.ResponseWriter, r *http.Request, owner string, collection string, key string, options *RequestOptions) {
	// GET wolk://
	if len(owner) == 0 {
		testnet_html, err := ioutil.ReadFile("/root/go/src/github.com/wolkdb/cloudstore/testnet/index.html")
		if err != nil {
			return
		}
		w.Write(testnet_html)
		return
	}

	// GET wolk://owner - return Collections
	if len(collection) == 0 {
		s.scanOwnerHandler(w, r, owner, options)
		return
	}
	//log.Info("[http:HTTPGetMethodHandler] going to getbucketobject", "owner", owner, "collection", collection)
	bucket, ok, _, err := s.GetBucketObject(owner, collection, options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		//TODO: [Rodney] - Check Account for ShimURL.  If present - try to retrieve from there
		http.Error(w, errBucketNotFound, http.StatusNotFound)
		return
	}
	//log.Info("[http:HTTPGetMethodHandler] bucket gotten", "b", fmt.Sprintf("%+v", bucket))
	if len(key) == 0 {
		s.scanCollectionHandler(w, r, owner, collection, options, bucket)
		return
	}

	//log.Info("[http:HTTPGetMethodHandler] key", "key", key, "owner", owner, "collection", collection)
	keyPieces := strings.Split(strings.Trim(key, "/"), "/")
	var index string
	if len(keyPieces) > 1 {
		index = keyPieces[0]
		key = keyPieces[1]
		//log.Info("[http:HTTPGetMethodHandler] keypices", "index", index, "key", key)
		s.getIndexedKeyHandler(w, r, owner, bucket.Name, index, key, options, bucket)
	} else {
		//GET wolk://owner/collection/key
		//log.Info("[http:HTTPGetMethodHandler] entering getKeyHandler, not indexed", "key", key)
		s.getKeyHandler(w, r, owner, bucket.Name, key, options, bucket)
	}
	return
}

//GetBucketObject -- Retrieve bucket corresponding to owner, bucket, etc ...
func (s *HttpServer) GetBucketObject(owner string, bucketName string, options *RequestOptions) (bucket *TxBucket, ok bool, proof *NoSQLProof, err error) {
	//log.Info("[http:GetBucketObject] begin", "o", owner, "b", bucketName)
	if options == nil {
		log.Error("[http:GetBucketObject] empty options... why?")
		options = NewRequestOptions()
	}
	txhash, ok, deleted, proof, err := s.wcs.GetBucket(owner, bucketName, options)
	if err != nil {
		log.Error("[http:GetBucketObject] GetKey", "owner", owner, "bucket", bucketName, "options", options.String(), "err", err)
		return bucket, ok, proof, err
	} else if !ok {
		log.Info("[http:GetBucketObject] NOT OK", "owner", owner, "bucket", bucketName, "options", options.String())
		return bucket, ok, proof, err
	} else if deleted {
		log.Info("[http:GetBucketObject] DELETED", "owner", owner, "bucket", bucketName, "options", options.String())
		return bucket, ok, proof, err
	}
	log.Info("[http:GetBucketObject] SUCC", "owner", owner, "bucket", bucketName, "options", options.String())
	tx, _, _, ok, err := s.wcs.GetTransaction(context.TODO(), txhash)
	if err != nil {
		log.Error("[http:GetBucketObject] GetTransaction", "owner", owner, "bucket", bucketName, "options", options.String())
		return bucket, ok, proof, err
	} else if !ok {
		return bucket, ok, proof, nil
	}
	bucket, err = tx.GetTxBucket()
	if err != nil {
		log.Error("[http:GetBucketObject] GetTxBucket", "owner", owner, "bucket", bucketName, "options", options.String())
		return bucket, true, proof, err
	}
	log.Info("[http:GetBucketObject] FINAL", "owner", owner, "bucket", fmt.Sprintf("%+v", bucket), "options", options.String())
	return bucket, true, proof, nil
}

//ParseHttpUrlPath -- Parse HTTP Url Path to retrieve owner, address and key
func (s *HttpServer) ParseHttpUrlPath(path string) (owner string, collection string, key string) {
	pathpieces := strings.Split(strings.Trim(path, "/"), "/")
	pathLength := len(pathpieces)
	if pathLength > 0 {
		owner = pathpieces[0]
		if pathLength > 1 {
			collection = pathpieces[1]
			if pathLength > 2 {
				key = strings.Join(pathpieces[2:], "/")
			}
		}
	}
	return owner, collection, key
}

func (s *HttpServer) GetSignerAddress(r *http.Request) (signer common.Address) {
	return crypto.GetSignerAddress(r.Header.Get("Requester"))
}

func (s *HttpServer) writeNoSQLScanProof(w http.ResponseWriter, p *NoSQLProof) {
	if p != nil {
		w.Header().Add("Access-Control-Expose-Headers", AccessControlExposeHeaders)
		log.Info("[http:writeNoSQLProof]", "proof", p)
		w.Header().Add("Proof-Type", "NoSQLScan")
		w.Header().Add("Proof", p.String())
	}
}

func (s *HttpServer) writeNoSQLProof(w http.ResponseWriter, p *NoSQLProof) {
	if p != nil {
		w.Header().Add("Access-Control-Expose-Headers", AccessControlExposeHeaders)
		log.Info("[http:writeNoSQLProof]", "proof", p)
		w.Header().Add("Proof-Type", "NoSQL")
		w.Header().Add("Proof", p.String())
	}
}

func (s *HttpServer) writeSMTProof(w http.ResponseWriter, p *SMTProof) {
	if p != nil {
		w.Header().Add("Access-Control-Expose-Headers", AccessControlExposeHeaders)
		log.Info("[http:writeSMTProof]", "proof", p)
		w.Header().Add("Proof-Type", "SMT")
		w.Header().Add("Proof", p.String())
	}
}

func (s *HttpServer) writeSimpleProof(w http.ResponseWriter, p *Proof) {
	if p != nil {
		w.Header().Add("Access-Control-Expose-Headers", AccessControlExposeHeaders)
		log.Info("[http:writeProof]", "proof", p)
		w.Header().Add("Proof-Type", "Simple")
		w.Header().Add("Proof", p.String())
	}
}

func (s *HttpServer) bandwidthHandler(w http.ResponseWriter, r *http.Request) {
	options := s.getRequestOptions(r)
	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var addr common.Address
	name_or_address_string := pathpieces[2]
	if len(name_or_address_string) < 32 {
		var ok bool
		var err error
		addr, ok, _, err = s.wcs.GetName(name_or_address_string, options)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if !ok {
			http.Error(w, errUserNotFound, http.StatusNotFound)
			return
		}
	} else {
		addr = common.HexToAddress(name_or_address_string)
	}

	balance, err := s.wcs.Storage.GetBandwidthBalance(addr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("%d", balance)))
}

func (s *HttpServer) getRequestOptions(r *http.Request) (options *RequestOptions) {
	options = NewRequestOptions()

	options.Proof = false
	proofFlag := r.Header.Get("Proof")
	if proofRequired, _ := strconv.ParseBool(proofFlag); proofRequired {
		options.Proof = proofRequired
	}

	// proof of replication request are made with a ReplicaChallenge header added to getKey requests
	replicaChallenge := r.Header.Get("ReplicaChallenge")
	if len(replicaChallenge) > 0 {
		challenge, err := strconv.Atoi(replicaChallenge)
		if err != nil {
			log.Error("[http:getRequestOptions] Error reading replicaChallenge ", "err", err)
		} else {
			log.Info("[http:getRequestOptions]", "replicaChallenge", challenge)
			options.ReplicaChallenge = int64(challenge)
		}
	} else {
		options.ReplicaChallenge = -1
	}

	blockNumNotSet := 0
	requestedBN := blockNumNotSet

	bn := r.Header.Get("Blocknumber")
	//log.Info("[http:getRequestOptions] received headers", "proof",  r.Header.Get("proof"), "Proof", r.Header.Get("Proof"), "User-Agent", r.Header.Get("User-Agent"))
	if len(bn) > 0 {
		passedInBlockNumber, err := strconv.Atoi(bn)
		if err != nil {
			log.Error("[http:getRequestOptions] Error reading blockNumber ", "err", err)
			return options
		}
		if passedInBlockNumber > 0 {
			requestedBN = passedInBlockNumber
		} else if passedInBlockNumber == LastFinalizedState || passedInBlockNumber == PreemptiveState {
			// -3, -2
			requestedBN = passedInBlockNumber
		} else if passedInBlockNumber == LastExternalState || passedInBlockNumber == LocalBestState {
			// -1, -4
			requestedBN = passedInBlockNumber
		} else {
			log.Error("[http:getRequestOptions] Invalid blocknumber passed in: ", "blocknumber", passedInBlockNumber)
			requestedBN = blockNumNotSet
		}

	} else {
		//BlockNumber Not Supplied, External 0
		requestedBN = blockNumNotSet
	}

	//to be removed, preemptive flag will overwrite proof and bn setting
	preemptiveFlag := r.Header.Get("IsPreemptive")
	ispreemptive, _ := strconv.ParseBool(preemptiveFlag)
	if ispreemptive {
		options.IsPreemptive = ispreemptive
		options.Proof = false
		options.BlockNumber = PreemptiveState
	}

	switch requestedBN {

	case blockNumNotSet:
		//Dynamically set BN
		if options.Proof == true {
			options.BlockNumber = LastFinalizedState
		} else {
			options.BlockNumber = PreemptiveState
		}

	case LastFinalizedState:
		//finalized request must have proof
		if options.Proof == false {
			log.Error("[http:getRequestOptions] Invalid RequestOptions: noProof + lastFinalized")
			options.Proof = true
		}
		options.BlockNumber = LastFinalizedState

	case PreemptiveState:
		//preemptive request can't ask for proof
		if options.Proof == true {
			log.Error("[http:getRequestOptions] Invalid RequestOptions: proof + preemptive")
			options.Proof = false
		}
		options.BlockNumber = PreemptiveState

	case LastExternalState:
		//proof is "optional"
		options.BlockNumber = LastConsensusState

	default:
		options.BlockNumber = requestedBN
	}

	maxContentLength := r.Header.Get("MaxContentLength")
	if len(maxContentLength) > 0 {
		n, err := strconv.Atoi(maxContentLength)
		if err != nil {
			return options
		} else {
			if n == 0 {
				options.MaxContentLength = uint64(DefaultMaxContentLength)
			} else {
				options.MaxContentLength = uint64(n)
			}
		}
	}

	waitfortx := r.Header.Get("WaitForTx")
	if len(waitfortx) > 0 {
		options.WaitForTx = waitfortx
	}

	indexes := r.Header.Get("Indexes")
	if len(indexes) > 0 {
		err := json.Unmarshal([]byte(indexes), &options.Indexes)
		if err != nil {
			log.Error("[http:getRequestOptions]", "err", err)
			panic(err) // TOOD: take out
		}
	}
	//log.Info("[http:getRequestOptions]", "options", options.String())
	return options
}

func (s *HttpServer) checkSigStub(w http.ResponseWriter, r *http.Request, payloadType string, body []byte) (options *RequestOptions, signer common.Address, balance uint64, bandwidth int64, statusCode int, err error) {
	options = s.getRequestOptions(r)
	return options, signer, 0, 0, 200, nil
}

// Here we check that the signature for the GET is good enough:
// 1. check the length
// 2. get the public key and verify that the signature matches the public KEY
// 3. get the signer address from the signature
// 4. get the WOLK balance using the address
// 5. ensure that the WOLK balance is greater than zero
// 6. get the bandwidth consumption level
/*
signing GET requests:  the string path
signing PUT requests:  8 byte actual size [visible to server] and then the string path.
signing Tx:            included in tx.Sig
*/

func (s *HttpServer) checkSig(w http.ResponseWriter, r *http.Request, payloadType string, body []byte) (options *RequestOptions, signer common.Address, balance uint64, bandwidth int64, statusCode int, err error) {

	options = s.getRequestOptions(r)

	// This is a JSON Web Key (JWK)
	requester := r.Header.Get("Requester")
	// This is a JSON Web Signature (JWS)
	sig := common.FromHex(r.Header.Get("Sig"))

	// TODO: take out PayloadBucket
	if payloadType == PayloadKey || payloadType == PayloadBucket || payloadType == PayloadBatch || payloadType == PayloadShareBatch || payloadType == PayloadChunk || payloadType == PayloadShare || payloadType == "info" || payloadType == "block" || payloadType == "tx" || payloadType == "sbatch" {
		// skip signature checks to focus on chunk availability
		//log.Debug(fmt.Sprintf("[http:checkSig] checkingSig for payloadType [%s] that has body of [%+v]", payloadType, body))
		return options, signer, 0, 0, 200, nil
	}
	msg := PayloadBytes(r, body)

	// bytes: JUST the path right Now but consider: (a) timestamp
	msgsupplied := r.Header.Get("Msg")
	verified, err := crypto.JWKVerifyECDSA(msg, requester, sig)
	if err != nil {
		w.Header().Set("WWW-Authenticate", `WOLK realm="WOLK"`)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("[http:checkSig] VerifySign", "path", r.URL.Path, "err", err, "payloadType", payloadType, "sig", string(sig), "msg", string(msg), "msgsupplied", string(msgsupplied))
		return options, signer, balance, bandwidth, http.StatusUnauthorized, err
	} else if !verified {
		http.Error(w, errVerificationFailure, http.StatusUnauthorized)
		log.Error("[http:checkSig] VerifySign STRING", "path", r.URL.Path, "err", err, "payloadType", payloadType, "sig", r.Header.Get("Sig"), "msg", string(msg), "msgsupplied", string(msgsupplied))
		log.Error("[http:checkSig] VerifySign HEX   ", "path", r.URL.Path, "err", err, "payloadType", payloadType, "sig", fmt.Sprintf("%x", sig), "msg", fmt.Sprintf("%x", msg), "msgsupplied", fmt.Sprintf("%x", msgsupplied))
		return options, signer, balance, bandwidth, http.StatusUnauthorized, err
	}
	signer = crypto.GetSignerAddress(requester)
	// log.Trace("[http:checkSig] VerifySign", "sig", sig, "requester", requester, "signer", signer)
	return options, signer, 0, 0, http.StatusOK, nil

	// look up the balances
	account, ok, err := s.wcs.GetAccount(signer, 0)
	if err != nil {
		log.Error("[http:checkSig] GetAccount", "path", r.URL.Path, "signer", signer, "err", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return options, signer, balance, bandwidth, http.StatusUnauthorized, err
	}
	// log.Info("[http:checkSig]", "balance", balance)
	if !ok { // address not found
		log.Error("[http:checkSig] Account NOT FOUND", "path", r.URL.Path, "signer", signer)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return options, signer, balance, bandwidth, http.StatusUnauthorized, fmt.Errorf("GetAccount NOT FOUND")
	}
	if account.Balance <= 0 {
		log.Error("[http:checkSig] NO BALANCE", "path", r.URL.Path, "signer", signer, "balance", balance)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return options, signer, balance, bandwidth, http.StatusUnauthorized, fmt.Errorf("GetBalance is ZERO")
	}
	// we are good!
	bandwidth, err = s.wcs.Storage.GetBandwidthBalance(signer)
	if err != nil {
		log.Error("checkSig-Bandwidth.GetBalance ERR", "path", r.URL.Path, "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return options, signer, balance, bandwidth, http.StatusInternalServerError, fmt.Errorf("GetSigner NOT FOUND")
	}
	log.Info("[http:checkSig] SUCC", "path", r.URL.Path, "signer", signer, "bandwidth", bandwidth, "balance", balance)
	return options, signer, balance, bandwidth, http.StatusOK, nil
}

func (s *HttpServer) setChunkHandler(w http.ResponseWriter, r *http.Request) {
	chunk, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	options, addr, balance, bandwidth, _, err := s.checkSig(w, r, PayloadChunk, chunk)
	if err != nil {
		return
	}
	log.Info("[http:setChunkHandler]", "options", options.String())

	var ctx context.Context
	/*
		tracer, closer := opentrace.Init("setChunkHandler")
		opentracing.SetGlobalTracer(tracer)
		defer closer.Close()

		spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		log.Trace("setChunkHandler", "spanCtx", spanCtx, "err", err)
		var span opentracing.Span

		//disable opentracing
		spanCtx = nil
		if spanCtx == nil {
			ctx = nil
		//	   tracer, closer := opentrace.Init("SetChunk")
		//	   defer closer.Close()
		//	   opentracing.SetGlobalTracer(tracer)

		//	   span = tracer.StartSpan("handler")
		//	   defer span.Finish()

		//	   log.Info("[http:setChunkHandler]", "span", span)
		} else {
			span = tracer.StartSpan("handler", opentracing.ChildOf(spanCtx))
			defer span.Finish()
			greeting := span.BaggageItem("greeting")
			if greeting == "" {
				greeting = "Hello"
			}

			span.LogFields(
				otlog.String("processing", "setChunkHandler"),
				otlog.String("value", greeting),
			)
			ext.SamplingPriority.Set(span, 1)

			ctx = opentracing.ContextWithSpan(context.Background(), span)
			log.Trace("setChunkHandler", "ctx", ctx)
			testspan, ctxtest := opentracing.StartSpanFromContext(ctx, "printHello")
			log.Trace("Handler", "testspan", testspan, "ctx", ctxtest)
		}
	*/

	// ENCODE
	k := wolkcommon.Computehash(chunk)

	h, err := s.wcs.Storage.QueueChunk(ctx, chunk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("setChunkHandler", "addr", addr, "k", fmt.Sprintf("%x", k))
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "{\"mr\":\"%x\",\"h\":\"%x\"}", h, k)
	log.Trace(fmt.Sprintf("setChunkHandler addr %x (%d|%d) chunkID %x err %v", addr, balance, bandwidth, h, err))
}

func (s *HttpServer) setShareHandler(w http.ResponseWriter, r *http.Request) {
	chunk, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	k := wolkcommon.Computehash(chunk)
	h := fmt.Sprintf("%x", k)
	options, addr, balance, bandwidth, _, err := s.checkSig(w, r, PayloadShare, chunk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Trace("[http:setShareHandler]", "options", options.String())

	var ctx context.Context
	/*
		tracer, closer := opentrace.Init("setChunkHandler")
		opentracing.SetGlobalTracer(tracer)
		defer closer.Close()

		spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		var span opentracing.Span
		//disable opentracing
		spanCtx = nil
		if spanCtx == nil {
			ctx = nil
		} else {
			span = tracer.StartSpan("handler", opentracing.ChildOf(spanCtx))
			defer span.Finish()
			greeting := span.BaggageItem("greeting")
			if greeting == "" {
				greeting = "Hello"
			}

			span.LogFields(
				otlog.String("processing", "setChunkHandler"),
				otlog.String("value", greeting),
			)
			log.Trace("Handler", "span", span)
			ext.SamplingPriority.Set(span, 1)

			ctx = opentracing.ContextWithSpan(context.Background(), span)
			log.Trace("setChunkHandler", "ctx", ctx)
			testspan, ctxtest := opentracing.StartSpanFromContext(ctx, "printHello")
			log.Trace("Handler", "testspan", testspan, "ctx", ctxtest)
		}
	*/
	merkleRoot, _, len, err := s.wcs.Storage.SetShare(ctx, chunk)
	if err == cloud.ErrWriteLimit {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(fmt.Sprintf("setShareHandler ID %x err %v", h, err))
	}
	// TODO: revisit this
	fmt.Fprintf(w, "{\"h\":\"%x\",\"mr\":\"%x\", \"len\":%d }", k, merkleRoot, len)
	log.Trace(fmt.Sprintf("setShareHandler addr %x (%d|%d) chunkID %x err %v", addr, balance, bandwidth, k, err))
}

type chunkSearch struct {
	Local  bool
	Remote bool
	Memory bool
	Val    []byte
}

func (s *HttpServer) getChunkSearchHandler(w http.ResponseWriter, r *http.Request) {
	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	h := pathpieces[2]
	if len(h) != 64 {
		http.Error(w, errIncorrectLength, http.StatusBadRequest)
		return
	}
	k, err := hex.DecodeString(h)
	log.Info("getChunkSearchHandler", "k", fmt.Sprintf("%x", k))
	var cs chunkSearch
	cs.Val, cs.Local, cs.Remote, cs.Memory, err = s.wcs.Storage.ChunkSearch(k)
	if err != nil {
		log.Error("[http:getChunkSearchHandler]", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	val, err := json.Marshal(cs)
	w.Write(val)
}

func (s *HttpServer) getChunkHandler(w http.ResponseWriter, r *http.Request) {
	options, addr, balance, bandwidth, _, err := s.checkSig(w, r, PayloadChunk, []byte{})
	log.Trace("[http:getChunkHandler]", "options", options.String())

	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	h := pathpieces[2]
	if len(h) != 64 {
		http.Error(w, errIncorrectLength, http.StatusBadRequest)
		return
	}
	k, err := hex.DecodeString(h)
	log.Info("getChunkHandler", "addr", addr, "k", fmt.Sprintf("%x", k))
	val, found, err := s.wcs.Storage.GetChunk(context.TODO(), k)
	if err != nil {
		log.Error("[http:getChunkHandler]", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !found {
		http.Error(w, errChunkNotFound, http.StatusNotFound)
		return
	}
	// w.Header().Set("Content-Type", "application/json")
	// w.Header().Set("Content-Disposition", "attachment; filename=x.jpeg")

	w.Write(val)
	s.TallyRequesterBandwidth(int64(len(val)), addr)
	log.Info(fmt.Sprintf("getChunkHandler addr %x (%d|%d) chunkID %x", addr, balance, bandwidth, h))
}

func isEventSQLRead(sqlReq *SQLRequest) bool {
	if IsReadQuery(sqlReq.RawQuery) || sqlReq.RequestType == http.MethodGet || sqlReq.RequestType == "ListDatabases" || sqlReq.RequestType == "DescribeTable" {
		return true
	}
	return false
}

//if there is no return error, shouldn't these errs be w.Write instead of just returning from the function?
func (server *HttpServer) sqlHandler(w http.ResponseWriter, r *http.Request, owner string, database string, options *RequestOptions) { //  b *TxBucket
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("[http:sqlHandler] read body", "err", err)
		return
	}
	log.Info("[http:sqlHandler]", "requestbody", string(body))
	sqlRequest := new(SQLRequest)
	err = json.Unmarshal(body, &sqlRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPatch {
		log.Info("[http:sqlHandler] SQL PATCH (ReadSQL)", "collection", r.URL.Path, "body", string(body))
		result, err := server.wcs.Read(sqlRequest, options)
		if err != nil {
			log.Error("[http:sqlHandler] Read", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Info("[http:sqlHandler] Read SUCC", "result", result)
		resultBytes, err := json.Marshal(result)
		if err != nil {
			log.Error("[http:sqlHandler] marshal result", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//w.Write([]byte(fmt.Sprintf("%s", result)))
		_, err = w.Write(resultBytes) // ingest int status of this?
		if err != nil {
			log.Error("[http:sqlHandler] w.Write result", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Info("[http:sqlHandler] read SUCCESS", "path", r.URL.Path)
	} else {
		log.Info("[http:sqlHandler] SQL POST (MutateSQL)", "collection", r.URL.Path, "body", string(body))
		server.submitTxHandler(w, r, PayloadSQL, body)
	}
}

func PayloadBytes(r *http.Request, body []byte) []byte {
	if r.Method == http.MethodPut {
		// This covers our base for now.
		return append([]byte(r.Method), []byte(r.URL.Path)...)
	}
	return append(append([]byte(r.Method), []byte(r.URL.Path)...), body...)
}

func (s *HttpServer) WaitForTx(txhash common.Hash) (waiterr error) {
	log.Info(fmt.Sprintf("[http:WaitForTx] START %x", txhash))
	txApplied := false
	numChecks := 0
	for txApplied == false {
		_, _, status, _, err := s.wcs.GetTransaction(context.TODO(), txhash)
		if err != nil {
			log.Error(fmt.Sprintf("[http:WaitForTx] ERROR: [%s] Get Transaction [%x]", err, txhash))
			return err
		}
		if status == "Applied" {
			log.Info(fmt.Sprintf("[http:WaitForTx] %x is DONE", txhash))
			txApplied = true
			break
		}
		log.Info(fmt.Sprintf("[http:WaitForTx] %x is WAITING", txhash))
		time.Sleep(time.Duration(5) * time.Second)
		numChecks++
		if numChecks > 24 {
			return fmt.Errorf("[http:WaitForTx] failed to return after 24 tries")
		}
	}
	return nil
}

func (s *HttpServer) createBucketHandler(w http.ResponseWriter, r *http.Request, owner string) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("[http:createBucketHandler]", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//log.Info("[http:createBucketHandler] bucket", "body", string(body))
	var b *TxBucket
	err = json.Unmarshal(body, &b)
	if err != nil {
		log.Error("[http:createBucketHandler] Unable to unmarshal bucket", "body", string(body))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	options, addr, balance, bandwidth, _, err := s.checkSig(w, r, PayloadBucket, body)
	if err != nil {
		log.Error("[http:createBucketHandler] cb2b", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// if this is a indexed bucket/collection, include the indexes in the bucket.
	if len(options.Indexes) > 0 {
		//log.Info("[http:createBucketHandler] options have indexes", "indexes", fmt.Sprintf("%+v", options.Indexes))
		for _, idx := range options.Indexes {
			b.Indexes = append(b.Indexes, idx)
		}
		body, err = json.Marshal(b)
		if err != nil {
			log.Error(fmt.Sprintf("[http:createBucketHandler] Unable to marshal bucket(%+v)", b))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//log.Info("[http:createBucketHandler] new body with indexes", "b", string(body))
	}

	var txhash common.Hash
	txhash, err = s.wcs.SendRawTransaction(NewTransactionImplicit(r, body))
	//log.Info("[http:createBucketHandler]", "txhash", fmt.Sprintf("%x", txhash), "err", err)
	if err != nil {
		log.Error(fmt.Sprintf("[http:createBucketHandler] SendRawTransaction - Error [%+v]", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if options.WaitForTx == "1" {
		waiterr := s.WaitForTx(txhash)
		if waiterr != nil {
			log.Error("[http:createBucketHandler] ERROR: WaitForTx", "error", waiterr)
			http.Error(w, waiterr.Error(), http.StatusInternalServerError)
			return
		}
	}

	log.Info("[http:createBucketHandler]", "path", r.URL.Path, "addr", addr, "balance", balance, "bandwidth", bandwidth)
	w.Write([]byte(fmt.Sprintf("%x", txhash)))
	return
}

func (s *HttpServer) setKeyHandler(w http.ResponseWriter, r *http.Request, owner string, collection string, key string, b *TxBucket) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("[http:setKeyHandler]", "err", err)
		return
	}
	sz := len(body)
	log.Info("[http:setKeyHandler] ", "sz", len(body))

	options, addr, balance, bandwidth, _, err := s.checkSig(w, r, PayloadKey, body)
	if err != nil {
		if b.ShimURL == "" {
			log.Info("sk2b", "err", err)
			return
		}
		//TODO: remove - just for testing shim
		err = nil
		bandwidth = 0
	}

	log.Info("[http:setKeyHandler]", "options", options.String(), "sz", sz)
	//reader := bytes.NewReader(body)
	ctx := context.Background()
	var fileHash common.Hash
	if useReplica {
		k, storageRoot, err := s.wcs.Storage.SetReplica(ctx, body)
		if err != nil {
			log.Error("[http:setKeyHandler] SetReplica", "err", err)
		} else {
			log.Info("[http:setKeyHandler] SetReplica", "k", k, "storageRoot", storageRoot)
		}
		fileHash = common.BytesToHash(k)
	} else {
		fileHashBytes, err := s.wcs.Storage.PutFile(ctx, body, nil)
		if err != nil {
			log.Error("[http:setKeyHandler] PutFile", "err", err)
		}
		fileHash = common.BytesToHash(fileHashBytes)
	}
	log.Info("[http:setKeyHandler] ", "fileHash", fileHash)
	if len(body) < 256 {
		//log.Info(fmt.Sprintf("[http:setKeyHandler] body == [%s]", body))
		//Comment for now to lower file size
	}
	// we have validated the signature already in checkSig
	//log.Info(fmt.Sprintf("[http:setKeyHandler] valid body == [%s]", body))
	txp := NewTxKey(fileHash, uint64(len(body)))
	tx := NewTransactionImplicit(r, []byte(txp.String()))
	log.Info("[http:setKeyHandler] NewTransactionImplicit", "txp", txp.String(), "sig", r.Header.Get("Sig"), "requester", r.Header.Get("Requester"), "msgsupplied", r.Header.Get("Msg"), "msgsupplied", fmt.Sprintf("%x", r.Header.Get("Msg")))
	//log.Info("[http:setKeyHandler] NewTransactionImplicit3", "tx", tx)
	_, err = s.wcs.SendRawTransaction(tx)
	if err != nil {
		log.Error(fmt.Sprintf("[http:setKeyHandler] SendRawTransaction - Error [%+v]", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("[http:setKeyHandler]", "path", r.URL.Path, "addr", addr, "balance", balance, "bandwidth", bandwidth, "sz", sz, "fileHash", fileHash)
	if r.Header.Get("SHIM") == "true" {
		//Don't return key if it's a shim write
		return
	}
	if options.WaitForTx == "1" {
		waiterr := s.WaitForTx(tx.Hash())
		if waiterr != nil {
			log.Error(fmt.Sprintf("[http:createBucketHandler] ERROR: %s WaitForTx %x", waiterr, tx.Hash()))
			http.Error(w, waiterr.Error(), http.StatusInternalServerError)
			return
		}
	}
	log.Info(fmt.Sprintf("[http:setKeyHandler] WaitForTx %x is RETURNING", tx.Hash()))
	w.Write([]byte(fmt.Sprintf("%x", tx.Hash())))
	return
}

func (s *HttpServer) setIndexedKeyHandler(writer http.ResponseWriter, req *http.Request, bucket *TxBucket) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("[http:setIndexedKeyHandler]", "err", err)
		return
	}
	sz := len(body)

	options, addr, balance, bandwidth, _, err := s.checkSig(writer, req, PayloadKey, body)
	if err != nil {
		if bucket.ShimURL == "" {
			log.Error("sk2b", "err", err)
			return
		}
		//TODO: remove - just for testing shim
		err = nil
		bandwidth = 0
	}
	//log.Info("[http:setIndexedKeyHandler]", "options", options.String(), "sz", sz)

	txns, err := MakeIndexedTransactionsImplicit(req, bucket, body)
	if err != nil {
		log.Error("[http:setIndexedKeyHandler] no indexed transactions made!", "err", err)
		return
	}

	for _, tx := range txns {
		_, err = s.wcs.SendRawTransaction(tx)
		if err != nil {
			log.Error("[http:setIndexedKeyHandler] SendRawTransaction", "err", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Info("[http:setIndexedKeyHandler]", "path", req.URL.Path, "addr", addr, "balance", balance, "bandwidth", bandwidth, "sz", sz)
		if req.Header.Get("SHIM") == "true" {
			//Don't return key if it's a shim write
			return
		}

		// TODO: could have optimization here: don't have to wait for each of these right now?
		if options.WaitForTx == "1" {
			waiterr := s.WaitForTx(tx.Hash())
			if waiterr != nil {
				log.Error(fmt.Sprintf("[http:setIndexedKeyHandler] ERROR: %s WaitForTx %x", waiterr, tx.Hash()))
				http.Error(writer, waiterr.Error(), http.StatusInternalServerError)
				return
			}
		}
		log.Info(fmt.Sprintf("[http:setIndexedKeyHandler] WaitForTx %x is RETURNING", tx.Hash()))
		writer.Write([]byte(fmt.Sprintf("%x\n", tx.Hash())))
		log.Info("[http:setIndexedKeyHandler]", "tx", tx.String(), "sig", req.Header.Get("Sig"), "requester", req.Header.Get("Requester"), "msgsupplied", req.Header.Get("Msg"), "msgsupplied", fmt.Sprintf("%x", req.Header.Get("Msg")))
	}

	return
}

func (s *HttpServer) scanOwnerHandler(w http.ResponseWriter, r *http.Request, owner string, options *RequestOptions) {
	options, addr, balance, bandwidth, _, err := s.checkSig(w, r, PayloadBucket, []byte{})
	if err != nil {
		return
	}
	log.Info("[http:scanOwnerHandler]", "options", options.String())

	maxLength := options.MaxContentLength
	txhashList, ok, proof, err := s.wcs.ScanCollection(owner, "", options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, errUserNotFound, http.StatusNotFound)
		return
	}
	sz := uint64(0)

	var buckets []*TxBucket
	for _, txhash := range txhashList {
		tx, bn, _, ok, err := s.wcs.GetTransaction(context.TODO(), txhash)
		if err != nil {
			log.Error("[http:scanOwnerHandler] GetTransaction", "txhash", txhash)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else if !ok {
			log.Error("[http:scanOwnerHandler] GetTransaction NOT OK", "txhash", txhash)
		} else {
			log.Info("[http:scanOwnerHandler] GetTransaction OK!", "tx", tx)
			if string(tx.Method) == http.MethodDelete {
				// TODO: create a way to UNDELETE or at least see the deleted item
				log.Info("[http:scanOwnerHandler] Deleted item", "key", tx.Collection())
			} else {
				b, err := tx.GetTxBucket()
				if err != nil {
					log.Error("[http:scanOwnerHandler] getVal ", "err", err)
				} else {
					buckets = append(buckets, b)
					sz += 128 // should be an overestimate
				}
				if options.WithProof() {
					stx := NewSerializedTransaction(tx)
					stx.BlockNumber = bn
					proof.AddScanProofTx(txhash, stx)
				}
			}
		}
		if sz > maxLength {
			break
		}
	}
	buckets_json, err := json.Marshal(buckets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if options.WithProof() {
		s.writeNoSQLScanProof(w, proof)
	}
	// https://stackoverflow.com/questions/33183071/golang-serialize-deserialize-an-empty-array-not-as-null
	if len(buckets) == 0 {
		w.Write([]byte("[]"))
	} else {
		w.Write(buckets_json)
	}
	log.Info("[http:scanOwnerHandler]", "path", r.URL.Path, "addr", addr, "balance", balance, "bandwidth", bandwidth)
}

func (s *HttpServer) scanCollectionHandler(w http.ResponseWriter, r *http.Request, owner string, collection string, options *RequestOptions, b *TxBucket) {
	options, addr, balance, bandwidth, _, err := s.checkSig(w, r, PayloadBucket, []byte{})
	log.Info("[http:scanCollectionHandler]", "options", options.String())

	maxLength := options.MaxContentLength
	txhashList, ok, proof, err := s.wcs.ScanCollection(owner, collection, options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, errUserNotFound, http.StatusNotFound)
		return
	}
	sz := uint64(0)
	var items []*BucketItem
	for _, txhash := range txhashList {
		tx, bn, _, ok, err := s.wcs.GetTransaction(context.TODO(), txhash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error("[http:scanCollectionHandler] GetTransaction ERR", "err", err)
		} else if !ok {
			// TODO
		} else {
			signer, err := tx.GetSignerAddress()
			if err != nil {
				log.Error("[http:scanCollectionHandler] GetSigner ERR", "err", err)
			} else {
				if string(tx.Method) == http.MethodDelete {
					// TODO: create a way to UNDELETE or at least see the deleted item
					log.Info("[http:scanCollectionHandler] Deleted item", "tx.Key", tx.Key())
				} else {
					txp, err := tx.GetTxKey()
					if err != nil {

					} else {
						item := new(BucketItem)
						item.Key = string(tx.Key())
						item.ValHash = txp.ValHash
						item.Size = txp.Amount
						item.Writer = signer
						items = append(items, item)
						sz += 128 // should be an overestimate
						log.Info("[http:scanCollectionHandler] Added item", "item.Key", item.Key)
					}
				}
			}
		}
		if sz > maxLength {
			log.Error("[http:scanCollectionHandler] exceeded size", "sz", sz, "maxLength", maxLength)
			break
		}
		if options.WithProof() {
			stx := NewSerializedTransaction(tx)
			stx.BlockNumber = bn
			proof.AddScanProofTx(txhash, stx)
		}
	}
	log.Info("[http:scanCollectionHandler] Got items", "len(items)", len(items))
	items_json, err := json.Marshal(items)
	if err != nil {
		log.Error("[http:scanCollectionHandler] Marshal", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if options.WithProof() {
		s.writeNoSQLScanProof(w, proof)
	}
	// https://stackoverflow.com/questions/33183071/golang-serialize-deserialize-an-empty-array-not-as-null
	if len(items) == 0 {
		w.Write([]byte("[]"))
	} else {
		w.Write(items_json)
	}
	sz = uint64(len(items_json))
	log.Info("[http:scanCollectionHandler]", "RESULT", string(items_json))

	s.TallyBandwidth(b, int64(sz), addr, owner)
	log.Info("[http:scanCollectionHandler]", "path", r.URL.Path, "addr", addr, "balance", balance, "bandwidth", bandwidth)
}

func (s *HttpServer) TallyBandwidth(b *TxBucket, sz int64, requester common.Address, owner string) {
	if b.RequesterPays > 0 {
		s.wcs.Storage.AddBandwidth(sz, requester)
	} else {
		var ownerAddr common.Address
		s.wcs.Storage.AddBandwidth(sz, ownerAddr)
	}
}

func (s *HttpServer) TallyRequesterBandwidth(sz int64, requester common.Address) {
	s.wcs.Storage.AddBandwidth(sz, requester)
}

// TODO: Implement ShimURL concept
func (s *HttpServer) shim(w http.ResponseWriter, r *http.Request, owner string, collection string, key string, options *RequestOptions, bucket *TxBucket) (val []byte, ok bool, err error) {
	if len(bucket.ShimURL) == 0 {
		log.Error("[http:shim] owner has no shim url", "owner", owner)
		return val, false, nil
	}
	shimURL := bucket.ShimURL
	log.Info(fmt.Sprintf("[http:shim] shimURL vs bucket.ShimURL -- %s vs %s", shimURL, bucket.ShimURL))
	url := fmt.Sprintf("%s/%s", shimURL, filepath.Join(collection, key)) //TODO: need to optimize based on where the ShimURL comes from -- fmt.Sprintf("%s/%s", shimURL, string(key))
	log.Info("[http:shim] Shim FETCH", "url", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error("[http:shim] Error doing shim request", "error", err)
		return val, false, fmt.Errorf("[backend_nosql:shim] ERROR executing shim request for [%s] -> [%s]", url, err)
	}

	httpclient := &http.Client{Timeout: time.Second * 15}
	resp, err := httpclient.Do(req)
	if err != nil {
		log.Error("[http:shim] FETCH", "err", err)
		return val, false, nil //RODNEY: made error == nil so it behaves as 404 fmt.Errorf("[http:shim] %s", err)
	}

	if resp.StatusCode != 200 {
		log.Debug(fmt.Sprintf("[http:shim] Shim Request Resulted in StatusCode [%d] and Status [%s]", resp.StatusCode, resp.Status))
		return val, false, nil
	}

	defer resp.Body.Close()
	reader, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("[http:shim] FETCH", "err", err)
		return val, false, fmt.Errorf("[http:shim] %s", err)
	}
	sz := len(reader)
	log.Info("[http:shim] FETCH SUCC Reader is: ", "reader", string(reader), "sz", sz)

	if sz == 0 {
		log.Info("[http:shim] Reader size = 0, so returning without doing setkey")
		return val, false, nil
	}

	//setkeyHandler
	setKeyReader := bytes.NewReader(reader)
	setKeyRequest, setKeyRequestErr := http.NewRequest(http.MethodPut, url, setKeyReader)
	if setKeyRequestErr != nil {
		log.Error("[http:shim] http.NewRequest error", "error", setKeyRequestErr)
	}

	//log.Info("[http:shim] SetKeyReader ", "skReader", string(setKeyReader))
	setKeyRequest.Header.Add("Sig", resp.Header.Get("Sig"))             //TODO: how does the wolk node sign
	setKeyRequest.Header.Add("Requester", resp.Header.Get("Requester")) //TODO: how to retrieve the wolknode's public key
	setKeyRequest.Header.Add("Msg", resp.Header.Get("Msg"))
	setKeyRequest.Header.Add("SHIM", "true")
	log.Info("[http:shim] Attempting to setKey", "path", r.URL.Path, "msg", resp.Header.Get("Msg"))
	s.setKeyHandler(w, setKeyRequest, owner, collection, key, bucket)
	return reader, true, nil
}

// Background: https://developer.mozilla.org/en-US/docs/Web/HTTP/Range_requests
// (supported) Single part ranges curl http://i.imgur.com/z4d4kWk.jpg -i -H "Range: bytes=0-1023"
// (NOT SUPPORTED) Multipart ranges  curl http://www.example.com -i -H "Range: bytes=0-50, 100-150"
func (s *HttpServer) getFileRangeOffset(r *http.Request) (startoffset int64, endoffset int64, rangerequest bool) {
	// Range: bytes=0-1023

	startoffset = int64(0)
	endoffset = int64(maxFileSize)
	rangerequest = false
	bytesRange := r.Header.Get("Range") // bytes=0-1023

	if len(bytesRange) > 0 && strings.Contains(bytesRange, "bytes=") {
		rg := strings.Replace(bytesRange, "bytes=", "", -1) // 0-1023

		//TODO: support multiple ranges -- i.e. bytes=0-50, 100-150
		r := strings.Split(rg, "-")
		var startoffseti int
		var endoffseti int
		var err error
		if len(r[0]) == 0 {
			startoffseti = 0
		} else {
			startoffseti, err = strconv.Atoi(r[0])
			if err != nil {
				return startoffset, endoffset, true //regardless of error, I believe we need this to be true
			}
		}

		if len(r[1]) == 0 { //to handle case of bytes=1000-
			endoffseti = -1
		} else {
			endoffseti, err = strconv.Atoi(r[1])
			if err != nil || endoffseti > int(maxFileSize) || endoffseti < 0 {
				return startoffset, endoffset, true
			}
		}

		rangerequest = true
		startoffset = int64(startoffseti)
		endoffset = int64(endoffseti)
	}
	return startoffset, endoffset, rangerequest
}

func (s *HttpServer) getKeyVal(w http.ResponseWriter, r *http.Request, owner string, collection string, key string, options *RequestOptions, bucket *TxBucket) (val []byte, ok bool, deleted bool, valHash common.Hash, proof *NoSQLProof, err error) {
	if options == nil {
		log.Error("[http:getKeyVal] empty options... why?")
		options = NewRequestOptions()
	}
	txhash, ok, deleted, proof, err := s.wcs.GetKey(owner, collection, key, options)
	if err != nil {
		log.Error("[http:getKeyVal] GetKey", "owner", owner, "collection", collection, "key", key, "options", options.String(), "err", err)
		return val, ok, deleted, valHash, proof, err
	} else if !ok {
		log.Info("[http:getKeyVal] NOT OK", "owner", owner, "collection", collection, "key", key, "options", options.String())
		log.Info("[http:getKeyVal] trying shim")
		val, ok, err = s.shim(w, r, owner, collection, key, options, bucket)
		if !ok {
			log.Info("[http:getKeyVal] tried shim and still NOT OK", "owner", owner, "collection", collection, "key", key, "options", options.String())
		} else {
			log.Info("[http:getKeyVal] SUCC tried shim and got result", "owner", owner, "collection", collection, "key", key, "options", options.String())
		}
		return val, ok, deleted, valHash, proof, err
	} else if deleted {
		log.Info("[http:getKeyVal] DELETED", "owner", owner, "collection", collection, "key", key, "options", options.String())
		return val, ok, deleted, valHash, proof, err
	}
	log.Info("[http:getKeyVal] SUCC", "owner", owner, "collection", collection, "key", key, "options", options.String())
	tx, _, _, ok, err := s.wcs.GetTransaction(context.TODO(), txhash)
	if err != nil {
		log.Error("[http:getKeyVal] GetTransaction", "owner", owner, "collection", collection, "key", key, "options", options.String())
		return val, ok, deleted, valHash, proof, err
	} else if !ok {
		return val, ok, deleted, valHash, proof, err //RODNEY: not an error, but ok=false fmt.Errorf("could not find collection %s", collection)
	}
	txp, err := tx.GetTxKey()
	if err != nil {
		log.Error("[http:getKeyVal] GetTxPayloadSetKey", "owner", owner, "collection", collection, "key", key, "options", options.String())
		return val, ok, deleted, valHash, proof, err
	}
	/*
		if txp.Amount > options.MaxContentLength {
			log.Info("[http:getKeyVal] MaxContentLength", "ownerAddr", ownerAddr, "collection", collection, "key", key, "options", options.String())
			return val, ok, deleted, valHash, proof, fmt.Errorf("Cannot return value (size %d > maxcontentlength %d)", txp.Amount, options.MaxContentLength)
		}
	*/
	valHash = txp.ValHash
	if options.WithReplicaChallenge() {
		replicaproof, err := s.wcs.Storage.GetReplicaProof(valHash.Bytes(), uint(options.ReplicaChallenge))
		if err != nil {
			return val, ok, deleted, valHash, proof, err
		}
		proof.ReplicaProof = replicaproof
		// we do not need to get the val, return right away
		return val, ok, deleted, valHash, proof, nil
	} else if options.WithProof() {
		proof.Tx = NewSerializedTransaction(tx)
	}
	startoffset, endoffset, rangerequest := s.getFileRangeOffset(r)
	val, err = s.getVal(valHash, startoffset, endoffset)
	if err != nil {
		return val, ok, deleted, valHash, proof, err
	}
	if rangerequest {
		// TODO: HTTP/1.1 206 Partial Content
		// write headers out: Content-Length is covered already
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", startoffset, endoffset, txp.Amount))
	}
	return val, ok, deleted, valHash, proof, nil
}

func (s *HttpServer) getIndexedKeyVal(w http.ResponseWriter, r *http.Request, owner string, collection string, index string, key string, options *RequestOptions, bucket *TxBucket) (val []byte, ok bool, deleted bool, valHash common.Hash, proof *NoSQLProof, err error) {

	// if options == nil {
	// 	log.Error("[http:getIndexedKeyVal] empty options... why?")
	// 	options = NewRequestOptions()
	// }
	idx, err := bucket.GetIndex(index)
	if err != nil {
		log.Error("[http:getIndexedKeyVal] index not gotten!", "err", err)
		return val, ok, deleted, valHash, proof, fmt.Errorf("[http:getIndexedKeyVal] %s", err)
	}
	keyBytes := idx.StringToBytes(key)
	txhash, ok, deleted, proof, err := s.wcs.GetIndexedKey(owner, collection, index, keyBytes, options)
	if err != nil {
		log.Error("[http:getIndexedKeyVal] GetIndexedKey", "owner", owner, "collection", collection, "key", key, "options", options.String(), "err", err)
		return val, ok, deleted, valHash, proof, fmt.Errorf("[http:getIndexedKeyVal] %s", err)
	} else if !ok {
		log.Error("[http:getIndexedKeyVal] NOT OK", "owner", owner, "collection", collection, "key", key, "options", options.String())
		// TODO: put shim back in:
		//log.Info("[http:getKeyVal] trying shim")
		// val, ok, err = s.shim(w, r, owner, collection, key, options, bucket)
		// if !ok {
		// 	log.Info("[http:getKeyVal] tried shim and still NOT OK", "owner", owner, "collection", collection, "key", key, "options", options.String())
		// } else {
		// 	log.Info("[http:getKeyVal] SUCC tried shim and got result", "owner", owner, "collection", collection, "key", key, "options", options.String())
		// }
		return val, false, deleted, valHash, proof, nil
	} else if deleted {
		log.Info("[http:getIndexedKeyVal] DELETED", "owner", owner, "collection", collection, "key", key, "options", options.String())
		return val, ok, true, valHash, proof, nil
	}

	log.Info("[http:getIndexedKeyVal] SUCC", "owner", owner, "collection", collection, "key", key, "options", options.String())

	// if txp.Amount > options.MaxContentLength {
	// 	log.Info("[http:getKeyVal] MaxContentLength", "ownerAddr", ownerAddr, "collection", collection, "key", key, "options", options.String())
	// 	return val, ok, deleted, valHash, proof, fmt.Errorf("Cannot return value (size %d > maxcontentlength %d)", txp.Amount, options.MaxContentLength)
	// }

	// TODO:
	// if options.WithProof() {
	// 	proof.Tx = NewSerializedTransaction(tx)
	// }

	record, valHash, ok, err := s.getRecord(txhash, idx)
	if err != nil {
		log.Error("[http:getIndexedKeyVal]", "err", err)
		return val, ok, deleted, valHash, proof, fmt.Errorf("[http:getIndexedKeyVal] %s", err)
	} else if !ok {
		log.Error("[http:getIndexedKeyVal] NOT OK", "owner", owner, "collection", collection, "key", key, "options", options.String())
		return val, false, deleted, valHash, proof, nil
	}
	return record, true, false, valHash, proof, nil
}

func (s *HttpServer) getVal(valHash common.Hash, startoffset int64, endoffset int64) (val []byte, err error) {
	if useReplica {
		var storageRoot []byte
		val, storageRoot, err = s.wcs.Storage.GetReplica(valHash.Bytes(), int(startoffset), int(endoffset))
		if err != nil {
			log.Error("[http:getVal] GetReplica", "err", err)
			return val, err
		}
		log.Info("[http:getVal] GetReplica", "valHash", valHash, "storageRoot", storageRoot)
		return val, nil
	}
	val, err = s.wcs.Storage.GetFileWithRange(context.TODO(), valHash.Bytes(), startoffset, endoffset)
	if err == io.EOF {
		log.Error("[http:getVal] GetFileWithRange ERROR EOF", "err", err)
		return val, nil
	} else if err != nil {
		log.Error("[http:getVal] GetFileWithRange ERROR Read", "err", err)
		return val, err
	}
	return val, nil
}

func (s *HttpServer) getRecord(txhash common.Hash, idx *BucketIndex) (record []byte, valHash common.Hash, ok bool, err error) {
	ctx := context.TODO()
	tx, _, _, ok, err := s.wcs.GetTransaction(context.TODO(), txhash)
	if err != nil {
		return record, valHash, false, fmt.Errorf("[http:getRecord] %s", err)
	} else if !ok {
		return record, valHash, false, nil
	}
	txKey, err := tx.GetTxKey()
	if err != nil {
		return record, valHash, false, fmt.Errorf("[http:getRecord] %s", err)
	}
	rec, ok, err := s.wcs.Storage.GetChunk(ctx, txKey.ValHash.Bytes())
	if err != nil {
		return record, valHash, false, fmt.Errorf("[http:getRecord] %s", err)
	} else if !ok {
		return record, valHash, false, nil
	}
	//log.Info("[http:getRecord]", "key", idx.Translate(txKey.Key), "rec gotten", string(rec))
	return rec, txKey.ValHash, true, nil
}

func (s *HttpServer) getKeyHistoryHandler(w http.ResponseWriter, r *http.Request, owner string, collection string, key string, options *RequestOptions, b *TxBucket) {
	log.Info("[http:getKeyHistoryHandler] START")
	options, addr, balance, bandwidth, _, err := s.checkSigStub(w, r, PayloadKey, []byte{})
	if err != nil {
		return
	}
	if b.RequesterPays > 0 {
		options, addr, balance, bandwidth, _, err = s.checkSig(w, r, PayloadKey, []byte{})
	}

	history, err := s.wcs.GetKeyHistory(owner, collection, key, options)
	if err != nil {
		log.Error("[http:getKeyHistoryHandler] GetKeyHistory", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("[http:getKeyHistoryHandler] DONE", "len(history)", len(history))

	history_json, err := json.Marshal(history)
	if err != nil {
		log.Error("[http:getKeyHistoryHandler] Marshal", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(history_json)
	sz := uint64(len(history_json))
	s.TallyBandwidth(b, int64(sz), addr, owner)
	log.Info("[http:getKeyHistoryHandler]", "path", r.URL.Path, "addr", addr, "balance", balance, "bandwidth", bandwidth, "sz", sz)
}

func (s *HttpServer) getKeyHandler(w http.ResponseWriter, r *http.Request, owner string, collection string, key string, options *RequestOptions, b *TxBucket) {
	var contentType string
	ext := filepath.Ext(key)
	switch ext {
	case ".dmg":
		contentType = "application/octet-stream"
	case ".htm", ".html":
		contentType = "text/html"
	case ".css":
		contentType = "text/css"
	case ".js":
		contentType = "application/javascript"
	default:
		//contentType = http.DetectContentType(output)
	}

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Sig, Msg, Requester, Proof, WaitForTx")

	urlValues := r.URL.Query()
	if len(urlValues.Get("history")) > 0 {
		s.getKeyHistoryHandler(w, r, owner, collection, key, options, b)
		return
	}

	options, addr, balance, bandwidth, _, err := s.checkSigStub(w, r, PayloadKey, []byte{})
	if b.RequesterPays > 0 {
		options, addr, balance, bandwidth, _, err = s.checkSig(w, r, PayloadKey, []byte{})
	}
	if err != nil {
		log.Error("[http:getKeyHandler] checkSig/Stub failure ", "err", err)
		return
	}
	log.Info("[http:getKeyHandler] calling getKeyVal", "owner", owner, "collection", collection, "key", key, "options", options.String())

	if options.BlockNumber > 0 {
		lastFinalizedBlockNumber := s.wcs.LastFinalized()
		if lastFinalizedBlockNumber < uint64(options.BlockNumber) {
			if options.Proof == true {
				log.Error("[http:getKeyHandler] requested block has not been finalized yet")
				http.Error(w, "Requested Block Hasn't been finalized", http.StatusConflict)
				return
			} else if options.Proof == false {
				log.Error("[http:getKeyHandler] requested block has not been finalized yet")
				http.Error(w, "Requested Block Not Found", http.StatusNotFound)
				return
			}
		}
	}

	output, ok, deleted, valHash, proof, err := s.getKeyVal(w, r, owner, collection, key, options, b)
	if err != nil {
		log.Error("[http:getKeyHandler] getKeyVal", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		log.Info("[http:getKeyHandler] getKeyVal NOT OK", "owner", owner, "collection", collection, "key", key, "options", options.String())
		http.Error(w, errKeyNotFound, http.StatusNotFound)
		return
	} else if deleted {
		log.Info("[http:getKeyHandler] getKeyVal DELETED", "owner", owner, "collection", collection, "key", key, "options", options.String())
		http.Error(w, errKeyDeleted, http.StatusNoContent)
		return
	}
	log.Info("[http:getKeyHandler] DONE", "len(val)", len(output))

	if contentType == "" {
		contentType = http.DetectContentType(output)
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(output)))
	if options.WithProof() {
		s.writeNoSQLProof(w, proof)
	}
	w.Write(output)
	s.TallyBandwidth(b, int64(len(output)), addr, owner)
	log.Info("[http:getKeyHandler]", "path", r.URL.Path, "addr", addr, "balance", balance, "bandwidth", bandwidth, "sz", len(output), "valHash", valHash)
}

func (s *HttpServer) getIndexedKeyHandler(w http.ResponseWriter, r *http.Request, owner string, collection string, index string, key string, options *RequestOptions, bucket *TxBucket) {
	// urlValues := r.URL.Query()
	// if len(urlValues.Get("history")) > 0 {
	// 	s.getKeyHistoryHandler(w, r, owner, collection, key, options, b)
	// 	return
	// }

	options, addr, balance, bandwidth, _, err := s.checkSigStub(w, r, PayloadKey, []byte{})
	if bucket.RequesterPays > 0 {
		options, addr, balance, bandwidth, _, err = s.checkSig(w, r, PayloadKey, []byte{})
	}
	if err != nil {
		log.Error("[http:getIndexedKeyHandler] check sig", "err", err)
		return
	}

	//log.Info("[http:getIndexedKeyHandler] calling getIndexedKeyVal", "owner", owner, "collection", collection, "key", key, "options", options.String())
	output, ok, deleted, valHash, _, err := s.getIndexedKeyVal(w, r, owner, collection, index, key, options, bucket)
	if err != nil {
		log.Error("[http:getIndexedKeyHandler] getIndexedKeyVal", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		log.Info("[http:getIndexedKeyHandler] getIndexedKeyVal NOT OK", "owner", owner, "collection", collection, "key", key, "options", options.String())
		http.Error(w, errKeyNotFound, http.StatusNotFound)
		return
	} else if deleted {
		log.Info("[http:getIndexedKeyHandler] getIndexedKeyVal DELETED", "owner", owner, "collection", collection, "key", key, "options", options.String())
		http.Error(w, errKeyDeleted, http.StatusNoContent)
		return
	}
	//log.Info("[http:getIndexedKeyHandler] DONE", "len(val)", len(output))

	// var contentType string
	// ext := filepath.Ext(key)
	// switch ext {
	// case ".htm", ".html":
	// 	contentType = "text/html"
	// case ".css":
	// 	contentType = "text/css"
	// case ".js":
	// 	contentType = "application/javascript"
	// default:
	contentType := http.DetectContentType(output)
	// }

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(output)))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Sig, Msg, Requester, Proof, WaitForTx")
	// if options.WithProof() {
	// 	s.writeNoSQLProof(w, proof)
	// }
	w.Write(output)
	s.TallyBandwidth(bucket, int64(len(output)), addr, owner)
	log.Info("[http:getKeyIndexedHandler]", "path", r.URL.Path, "addr", addr, "balance", balance, "bandwidth", bandwidth, "sz", len(output), "valHash", valHash)
}

func (s *HttpServer) deleteKeyHandler(w http.ResponseWriter, r *http.Request, owner string, collection string, key string) {
	log.Info("[http:deleteKeyHandler]", "owner", owner, "collection", collection, "key", key)
	options, addr, balance, bandwidth, _, err := s.checkSig(w, r, PayloadKey, []byte{})
	log.Trace("[http:deleteKeyHandler]", "options", options.String())

	txhash, err := s.wcs.SendRawTransaction(NewTransactionImplicit(r, []byte{}))
	if err != nil {
		log.Error("[http:deleteKeyHandler]", "err", err)
	}
	log.Info("[http:deleteKeyHandler]", "path", r.URL.Path, "addr", addr, "balance", balance, "bandwidth", bandwidth)
	w.Write([]byte(fmt.Sprintf("%x", txhash)))
	return
}

func (s *HttpServer) getShareHandler(w http.ResponseWriter, r *http.Request) {
	options, addr, balance, bandwidth, _, err := s.checkSig(w, r, PayloadShare, []byte{})
	if fullTracing {
		log.Trace("[http:getShareHandler]", "options", options.String())
	}
	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	h := pathpieces[2]
	if len(h) != 64 {
		http.Error(w, errIncorrectLength, http.StatusBadRequest)
		return
	}
	k, err := hex.DecodeString(h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	val, len_chunk, err := s.wcs.Storage.GetShare(k)
	if err == cloud.ErrReadLimit {
		http.Error(w, err.Error(), http.StatusGatewayTimeout)
		return
	}
	if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Chunk-Len", fmt.Sprintf("%d", len_chunk))
	w.Write(val)
	s.TallyRequesterBandwidth(int64(len(val)), addr)
	if fullTracing {
		log.Trace("[http:getShareHandler]", "path", r.URL.Path, "addr", addr, "balance", balance, "bandwidth", bandwidth)
	}
}

func StrContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

func (s *HttpServer) setBatchHandler(w http.ResponseWriter, r *http.Request) {
	chunkData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	options, addr, balance, bandwidth, _, err := s.checkSig(w, r, PayloadBatch, chunkData)
	log.Trace("[http:setBatchHandler]", "options", options.String())
	/*
		var ctx context.Context
			tracer, closer := opentrace.Init("setChunkBatchHandler")
			opentracing.SetGlobalTracer(tracer)
			defer closer.Close()

			spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
			log.Info("[http:setChunkBatchHandler]", "spanCtx", spanCtx, "err", err)
			var span opentracing.Span
			//disable opentracing
			spanCtx = nil
			if spanCtx == nil {
				ctx = nil
			} else {
				span = tracer.StartSpan("handler", opentracing.ChildOf(spanCtx))
				defer span.Finish()
				greeting := span.BaggageItem("greeting")
				if greeting == "" {
					greeting = "Hello"
				}

				span.LogFields(
					otlog.String("processing", "setChunkHandler"),
					otlog.String("value", greeting),
				)
				log.Trace("Handler", "span", span)
				ext.SamplingPriority.Set(span, 1)

				ctx = opentracing.ContextWithSpan(context.Background(), span)
				log.Trace("setChunkHandler", "ctx", ctx)
				testspan, ctxtest := opentracing.StartSpanFromContext(ctx, "printHello")
				log.Trace("Handler", "testspan", testspan, "ctx", ctxtest)
			}
	*/

	// ENCODE
	var chunks []*cloud.RawChunk
	if err := json.Unmarshal(chunkData, &chunks); err != nil {
		return
	}
	//err = s.wcs.Storage.SetChunkBatch(ctx, chunks)
	for _, chunk := range chunks {
		s.wcs.Storage.AddQueue(chunk)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "text/plain")
	//fmt.Fprintf(w, "{\"mr\":\"%x\",\"h\":\"%x\"}", h, chunkKey)
	fmt.Fprintf(w, "{\"err\":\"%v\"}", err)
	log.Trace("[http:setChunkBatchHandler]", "addr", addr, "balance", balance, "bandwidth", bandwidth)
}

func (s *HttpServer) setShareBatchHandler(w http.ResponseWriter, r *http.Request) {
	//start := time.Now()
	chunkData, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	//options, addr, balance, bandwidth, _, err := s.checkSig(w, r, PayloadShare, chunkData)

	var shares []*cloud.RawChunk
	if err := json.Unmarshal(chunkData, &shares); err != nil {
		log.Error("[http:setShareBatchHandler] json.Unmarshal ", "err", err, "string(chunkData)", fmt.Sprintf("%s", chunkData), "len(chunkData)", len(chunkData), "chunkData", chunkData)
		return
	}
	ctx := context.Background()
	err = s.wcs.Storage.SetShareBatch(ctx, shares)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "text/plain")
	//fmt.Fprintf(w, "{\"mr\":\"%x\",\"h\":\"%x\"}", h, chunkKey)
	fmt.Fprintf(w, "{\"err\":\"%v\"}", err)
	//log.Info("[http:setShareBatchHandler]", "addr", addr, "balance", balance, "bandwidth", bandwidth)
	//log.Info("[http:setShareBatchHandler] finish", "len", len(shares), "time", time.Since(start))
}

func (s *HttpServer) getBatchHandler(w http.ResponseWriter, r *http.Request) {
	var chunks []*cloud.RawChunk
	chunkData, err := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(chunkData, &chunks); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	options, addr, balance, bandwidth, _, err := s.checkSig(w, r, PayloadBatch, chunkData)
	if err != nil {
		return
	}
	log.Trace("[http:getBatchHandler]", "options", options.String())

	/*
		   	tracer, closer := opentrace.Init("BatchHandler")
		   	opentracing.SetGlobalTracer(tracer)
		   	defer closer.Close()

		   	spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		   	log.Info("[http:getChunkBatchHandler]", "spanCtx", spanCtx, "err", err)
		   	var span opentracing.Span
		   	var ctx context.Context
		   	//disable opentracing
		   	spanCtx = nil
		   	if spanCtx == nil {
		   		ctx = nil
		   	//	   tracer, closer := opentrace.Init("SetChunk")
		   //		   defer closer.Close()
		   //		   opentracing.SetGlobalTracer(tracer)
		   //
		   //		   span = tracer.StartSpan("handler")
		   //		   defer span.Finish()
		   //
		 //		   log.Info("[http:getChunkBatchHandler]", "span", span)
		   	} else {
		   		span = tracer.StartSpan("handler", opentracing.ChildOf(spanCtx))
		   		defer span.Finish()
		   		greeting := span.BaggageItem("greeting")
		   		if greeting == "" {
		   			greeting = "Hello"
		   		}

		   		span.LogFields(
		   			otlog.String("processing", "getChunkBatchHandler"),
		   			otlog.Int("value", len(chunks)),
		   		)
		log.Info("[http:getChunkBatchHandler]", "span", span)
		   		ext.SamplingPriority.Set(span, 1)

		   		ctx = opentracing.ContextWithSpan(context.Background(), span)
		   		log.Trace("setChunkHandler", "ctx", ctx)
		   		testspan, ctxtest := opentracing.StartSpanFromContext(ctx, "printHello")
		   		log.Trace("Handler", "testspan", testspan, "ctx", ctxtest)
		   	}
	*/

	respChunks, err := s.wcs.Storage.GetChunkBatch(chunks)
	/*
		if spanCtx != nil {
			span.LogFields(
				otlog.String("processing", "ret from GetChunkBatch"),
				otlog.String("duration", batchtime.String()),
				otlog.Int("ret", len(respChunks)),
			)
		}
	*/

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("[http:getChunkBatchHandler]", "err", err)
		return
	}
	jchunks, err := json.Marshal(respChunks)

	log.Info(fmt.Sprintf("[http:getChunkBatchHandler] ID len(res)=%d", len(respChunks)))
	// check header
	w.Header().Set("Content-Type", "text/plain")
	w.Write(jchunks)
	s.TallyRequesterBandwidth(int64(len(jchunks)), addr)
	log.Info("[http:setChunkBatchHandler]", "addr", addr, "balance", balance, "bandwidth", bandwidth)
}

// /wolk/tx - POST
func (s *HttpServer) submitTxHandler(w http.ResponseWriter, r *http.Request, PayloadType string, body []byte) {
	var tx Transaction
	tx.Method = []byte(r.Method)
	tx.Path = []byte(r.URL.Path)
	tx.Payload = body
	tx.Sig = common.FromHex(r.Header.Get("Sig"))
	tx.Signer = []byte(r.Header.Get("Requester"))
	log.Info("[http:submitTxHandler]", "len(r.Header.Get(Sig))", len(r.Header.Get("Sig")), "len(sig)", len(tx.Sig), "len(body)", len(body), "tx", tx)
	validated, err := tx.ValidateTx()
	if err != nil {
		log.Error("[http:submitTxHandler] validatetx", "tx", tx.Hash(), "err", err)
		return // why was this not here?
	} else if !validated {
		log.Error("[http:submitTxHandler] NOT VALIDATED -- who cares, keep going", "tx", tx.Hash())
		http.Error(w, errInvalidSigner, http.StatusBadRequest)
	} else {
		log.Info("[http:submitTxHandler] VALIDATED", "tx", tx.Hash())
	}
	txhash, err := s.wcs.SendRawTransaction(&tx)
	if err != nil {
		log.Error("[http:submitTxHandler] txn errored!", "tx", txhash, "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("[http:submitTxHandler] txn SUBMIT SUCCESS", "tx", txhash)
	w.Write([]byte(fmt.Sprintf("%x", txhash)))
}

// getTxHandler
func (s *HttpServer) getTxHandler(w http.ResponseWriter, r *http.Request) {
	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathpieces) < 3 {
		http.Error(w, errInvalidTxRequest, http.StatusBadRequest)
		return
	}

	txhash_string := pathpieces[2]
	if len(txhash_string) < 64 {
		http.Error(w, errIncorrectLength, http.StatusBadRequest)
		return
	}

	var tx *Transaction
	var ok bool
	txhash := common.HexToHash(txhash_string)
	tx, blockNumber, txStatus, ok, err := s.wcs.GetTransaction(context.TODO(), txhash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, errTxNotFound, http.StatusNotFound)
		return
	}
	stx := NewSerializedTransaction(tx)
	stx.BlockNumber = blockNumber
	stx.Status = txStatus
	signer, err := tx.GetSignerAddress()
	if err != nil {

	} else {
		stx.SignerAddress = signer
	}
	w.Write([]byte(stx.String()))
}

func (s *HttpServer) getAccountHandler(w http.ResponseWriter, r *http.Request) {
	options := s.getRequestOptions(r)
	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(pathpieces) < 3 {
		http.Error(w, errInvalidAccountRequest, http.StatusInternalServerError)
		return
	}
	var address common.Address
	name_or_address_string := pathpieces[2]
	if len(name_or_address_string) < 32 {
		addr, ok, _, err := s.wcs.GetName(name_or_address_string, options)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if !ok {
			http.Error(w, errUserNotFound, http.StatusNotFound)
			return
		}
		address = addr
	} else {
		address = common.HexToAddress(name_or_address_string)
	}

	account, ok, err := s.wcs.GetAccount(address, options.BlockNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, errUserNotFound, http.StatusNotFound)
		return
	}
	w.Write([]byte(fmt.Sprintf("%s", account.String())))
}

func (s *HttpServer) getBlockHandler(w http.ResponseWriter, r *http.Request) {
	blockNumber := uint64(0)

	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathpieces) >= 3 {
		if pathpieces[2] == "latest" {
			blockNumber, _ = s.wcs.LatestBlockNumber()
			log.Info("getBlockHandler latest", "blockNumber", blockNumber)
			w.Write([]byte(fmt.Sprintf("%d", blockNumber)))
			return
		} else if pathpieces[2] == "lastfinalized" {
			blockNumber = s.wcs.LastFinalized()
			log.Info("getBlockHandler lastfinalized", "blockNumber", blockNumber)
			w.Write([]byte(fmt.Sprintf("%d", blockNumber)))
			return
		}
		bn, err := strconv.Atoi(pathpieces[2])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if bn > 0 {
			blockNumber = uint64(bn)
		}
	}

	b, ok, err := s.wcs.GetBlockByNumber(blockNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, errBlockNotFound, http.StatusNotFound)
		return
	}
	w.Write([]byte(b.String()))
}

func (s *HttpServer) getVoteHandler(w http.ResponseWriter, r *http.Request) {
	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathpieces) >= 3 {
		bn, err := strconv.Atoi(pathpieces[2])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		eng := s.wcs.protocolManager.getConsensusEngine(uint64(bn))
		if eng != nil {
			vm, _, _ := eng.getVoteLastStep()
			if vm != nil {
				w.Write(vm.Bytes())
				return
			} else {
				err := fmt.Errorf("No round found in getVoteLastStep[%d] ==> step %d", bn, eng.step)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			err := fmt.Errorf("No round found in getConsensusEngine [%d]", bn)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	err := fmt.Errorf("No round found in url")
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

type wolkInfo struct {
	Peers    int      `json:"peers"`
	PeerList []string `json:"peerlist"`

	ConsensusStatus                  string `json:"consensusStatus"`
	ConsensusStatusUpdatedSecondsAgo int64  `json:"consensusStatusUpdatedSecondsAgo"`
	Step                             uint64 `json:"step"`
	NetworkLatency                   uint64 `json:"networkLatency"`
	ExpectedTentative                int    `json:"expectedTentative"`
	ExpectedFinal                    int    `json:"expectedFinal"`
	TokenWeight                      uint64 `json:"tokenWeight"`

	IsProposer  bool   `json:"isProposer"`
	IsVoting    uint64 `json:"isVoting"`
	IsMalicious uint64 `json:"isMalicious"`

	HasBlock   bool   `json:"hasBlock"`
	Reserved   string `json:"reserved"`
	CountVotes uint64 `json:"countVotes"`

	BlockNumber               uint64                `json:"blockNumber"`
	Alloc                     uint64                `json:"alloc"`
	TotalAlloc                uint64                `json:"totalAlloc"`
	Sys                       uint64                `json:"sys"`
	NumGoroutine              int                   `json:"numGoroutine"`
	MLog                      MemLog                `json:"memlog"`
	StateDBCacheSize          int                   `json:"stateDBCacheSize"`
	ChunkCacheSize            int                   `json:"chunkCacheSize"`
	ConsensusAlgorithm        string                `json:"consensusAlgorithm"`
	IsPreemptive              bool                  `json:"isPreemptive"`
	LastFinalized             uint64                `json:"lastFinalizedRound"`
	LastFinalizedHash         common.Hash           `json:"lastFinalizedHash"`
	LastBlockMintedSecondsAgo int64                 `json:"lastBlockMintedSecondsAgo"`
	FetchChainLength          int                   `json:"fetchChainLength"`
	InsertChainLength         int                   `json:"insertChainLength"`
	PendingTxCount            int64                 `json:"pendingTxCount"`
	NumConnections            int64                 `json:"numConnections"`
	NetworkID                 uint64                `json:"networkID,omitempty"`
	BuildInfoSum              string                `json:"buildInfoSum"`
	Certs                     []*certificateSummary `json:"certs"`
}

type MemLog struct {
	Alloc       uint64
	NextGCLvl   uint64
	TotalAlloc  uint64
	Sys         uint64
	Mallocs     uint64
	Frees       uint64
	LiveObjects uint64

	PauseTotalNs uint64
	NumGC        uint32
	GCSys        uint64
	LastGC       uint64
	Age          int64 //in nano second
}

func getMemLog() MemLog {
	var m MemLog
	var rtm runtime.MemStats
	// Read full mem stats
	runtime.ReadMemStats(&rtm)

	// Misc memory stats
	m.Alloc = rtm.Alloc //HeapAlloc
	m.NextGCLvl = rtm.NextGC
	m.TotalAlloc = rtm.TotalAlloc
	m.Sys = rtm.Sys
	m.Mallocs = rtm.Mallocs
	m.Frees = rtm.Frees

	// Live objects = Mallocs - Frees
	m.LiveObjects = m.Mallocs - m.Frees

	// GC Stats
	m.PauseTotalNs = rtm.PauseTotalNs
	m.NumGC = rtm.NumGC
	m.GCSys = rtm.GCSys
	m.LastGC = rtm.LastGC / uint64(time.Millisecond)
	m.Age = (time.Now().UnixNano() / int64(time.Millisecond)) - int64(m.LastGC)
	return m
}

func buildInfoSum() string {
	if strings.HasSuffix(os.Args[0], ".test") {
		return "In Test Mode"
	} else {
		execfn, err := os.Executable()
		if err != nil {
			log.Error("[http:buildInfoSum]", "err", err)
			return fmt.Sprintf("%v", err)
		}
		if fs, err := os.Stat(execfn); err == nil {
			return fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", fs.Size()))))
		} else {
			log.Error("[http:buildInfoSum] Stat", "err", err)
			return fmt.Sprintf("%v", err)
		}
	}
}

func (s *HttpServer) getInfoHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	var info wolkInfo
	info.Peers = s.wcs.NumPeers()
	info.PeerList = s.wcs.GetPeerList()

	info.BlockNumber, _ = s.wcs.LatestBlockNumber()
	info.Alloc = m.Alloc
	info.TotalAlloc = m.TotalAlloc
	info.Sys = m.Sys
	info.NumGoroutine = runtime.NumGoroutine()
	info.MLog = getMemLog()
	info.LastFinalized = s.wcs.LastFinalized()
	info.LastFinalizedHash = s.wcs.LastFinalizedHash()
	info.LastBlockMintedSecondsAgo = s.wcs.GetLastBlockMintedSecondsAgo()
	info.ConsensusAlgorithm = "algo"
	info.StateDBCacheSize = s.wcs.Storage.StateDBCacheSize()
	info.ChunkCacheSize = s.wcs.Storage.ChunkCacheSize()
	info.PendingTxCount = s.wcs.GetPendingTxCount()

	info.NetworkID = s.wcs.genesis.NetworkID
	info.NumConnections = handlerCount()
	info.BuildInfoSum = buildInfoSum()
	info.IsPreemptive = s.wcs.GetIsPremptive()
	info.NetworkLatency = NetworkLatency
	info.ConsensusStatus, info.ConsensusStatusUpdatedSecondsAgo, info.Step,
		info.IsVoting, info.IsMalicious, info.ExpectedTentative, info.ExpectedFinal, info.TokenWeight,
		info.HasBlock, info.Reserved = s.wcs.getConsensusStatus(s.wcs.currentRound())

	certs, err := s.wcs.Indexer.getCertificateSummary(info.BlockNumber)
	if err == nil {
		info.Certs = certs
	}
	str, _ := json.Marshal(info)
	w.Write([]byte(str))
}

func (s *HttpServer) getGenesisHandler(w http.ResponseWriter, r *http.Request) {
	genesisFile := s.wcs.GenesisFile()
	genesisData, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(genesisData))
}

func (s *HttpServer) getNodeHandler(w http.ResponseWriter, r *http.Request) {
	nodenumber := s.wcs.GetIndex()
	options := s.getRequestOptions(r)
	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathpieces) >= 3 {
		nn, err := strconv.Atoi(pathpieces[2])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		nodenumber = int(nn)
	}

	var n *RegisteredNode
	n, ok, proof, err := s.wcs.GetNode(uint64(nodenumber), options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, errNodeNotFound, http.StatusNotFound)
		return
	}
	s.writeSimpleProof(w, proof)
	w.Write([]byte(n.String()))
}

func (s *HttpServer) getNameHandler(w http.ResponseWriter, r *http.Request) {
	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathpieces) < 3 || len(pathpieces[2]) == 0 {
		http.Error(w, errInvalidNameRequest, http.StatusBadRequest)
		return
	}
	options := s.getRequestOptions(r)

	name := pathpieces[2]
	addr, ok, proof, err := s.wcs.GetName(name, options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, errUserNotFound, http.StatusNotFound)
		return
	}
	s.writeSMTProof(w, proof)
	w.Write([]byte(fmt.Sprintf("%x", addr)))
}

func (s *HttpServer) usageHandler(w http.ResponseWriter, r *http.Request) {
	output := `Usage:\n
/wolk/latest/blocknumber\n
/wolk/balance/address\n
/wolk/tx/txhash\n
/wolk/block/blocknumber\n
/wolk/node/nodenumber\n
/wolk/name/name`
	w.Write([]byte(output))
}
