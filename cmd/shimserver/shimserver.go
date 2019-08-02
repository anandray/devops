package main

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
	//"context"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk"
)

var baseurl = "/root/go/src/github.com/wolkdb/cloudstore/cmd/shimserver/"

type HttpServer struct {
	privateKey  *ecdsa.PrivateKey
	pubKey      string
	Handler     http.Handler
	HTTPPort    int
	connections chan struct{}
}

const (
	maxConns = 1234
)

func (s *HttpServer) getConnection(w http.ResponseWriter, r *http.Request) {
	select {
	case <-s.connections:
		defer s.releaseConnection()
		s.handler(w, r)
	default:
		http.Error(w, "503 Service Unavailable", http.StatusServiceUnavailable)
		log.Error("[http:getConnection] 503 Service Unavailable", "goroutine", runtime.NumGoroutine())
	}
}

func (s *HttpServer) releaseConnection() {
	s.connections <- struct{}{}
}

func main() {
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))

	privKey := `{"kty":"EC","crv":"P-256","x":"d22uwkAD9ou144x1m_gLEEjytXyxokHgVnjMQhENxjs","y":"_fX9VT4KPsbl81HrcomYqaOHy1RMfaq7BB-nCnyqFfI","d":"rwPquDfRZZOuqzlatMjIsSHHVnitJZyndFRTZMGcc7U"}`
	pubKey := `{"kty":"EC","crv":"P-256","x":"d22uwkAD9ou144x1m_gLEEjytXyxokHgVnjMQhENxjs","y":"_fX9VT4KPsbl81HrcomYqaOHy1RMfaq7BB-nCnyqFfI"}`
	privateKey, err := crypto.JWKToECDSA(privKey)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(0)
	}

	s := &HttpServer{
		connections: make(chan struct{}, maxConns),
		pubKey:      pubKey,
		privateKey:  privateKey,
		HTTPPort:    99,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.getConnection)
	s.Handler = mux
	for i := 0; i < maxConns; i++ {
		s.connections <- struct{}{}
	}

	srv := &http.Server{
		Handler:      s.Handler,
		ReadTimeout:  600 * time.Second,
		WriteTimeout: 600 * time.Second,
	}
	/*
		if len(config.SSLCertFile) > 0 && len(config.SSLKeyFile) > 0 {
			err := srv.ListenAndServeTLS(config.SSLCertFile, config.SSLKeyFile)
			if err != nil {
				log.Error("[http:Start] ListenAndServeTLS error", "err", err)
			}
			log.Info("[http:Start] ListenAndServeTLS", "port", s.HTTPPort)
		}
	*/
	fmt.Printf("Listening on %d", s.HTTPPort)
	srv.Addr = fmt.Sprintf(":%d", s.HTTPPort)
	err = srv.ListenAndServe()
	if err != nil {
		log.Error("[http:Start] ListenAndServe error", "err", err)
	}
	log.Info("[http:Start] ListenAndServe", "port", s.HTTPPort)

}

// c0.wolk.com/archive.org/
func (s *HttpServer) handler(w http.ResponseWriter, r *http.Request) {
	if strings.TrimLeft(r.URL.Path, "/") == "healthcheck" {
		fmt.Fprintf(w, "%s", "OK")
	}
	url := fmt.Sprintf("https:/%s", r.URL.Path)
	// 1. map r.URL.Path onto a local directory, or another url outside ...
	//ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	//defer cancel()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	//	req.Cancel = ctx.Done()
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(0)
	}
	httpclient := &http.Client{Timeout: time.Second * 5} // , Transport: DefaultTransport}
	resp, err := httpclient.Do(req)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("readall %v", err)
		return
	}
	resp.Body.Close()
	statuscode := resp.StatusCode

	if statuscode == http.StatusNotFound {
		//w.NotFound()
		fmt.Printf("Not found")
		return
	}
	fmt.Printf("%s\n", body)
	// 2. sign it with the PrivateKey, return Sig + Requester headers
	// TODO: get the payload bytes right.

	returnReq, err := http.NewRequest(http.MethodPut, "http://shim"+r.URL.String(), nil)
	msg := wolk.PayloadBytes(returnReq, body)
	//fmt.Printf("\nMSG = %+s ReturnReq %+v RRURL %s \n", msg, returnReq, returnReq.URL.String())
	sig, err := crypto.JWKSignECDSA(s.privateKey, msg)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	w.Header().Add("Sig", fmt.Sprintf("%x", sig))
	w.Header().Add("Requester", s.pubKey)
	w.Header().Add("Msg", fmt.Sprintf("%s", msg))
	fmt.Fprintf(w, "%s", body)
}
