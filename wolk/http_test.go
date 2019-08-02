package wolk

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	//	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/crypto"
	filetype "gopkg.in/h2non/filetype.v1"
)

type MiniServer struct {
	Handler  http.Handler
	HTTPPort int
}

func (s *MiniServer) getConnection(w http.ResponseWriter, r *http.Request) {
	// do something amazing
	w.Header().Set("Access-Control-Allow-Origin", "*")
	auth := r.Header.Get("Authorization")
	proofHeader := r.Header.Get("Proof")
	msgHeader := r.Header.Get("Msg")
	Requester := r.Header.Get("Requester")
	log.Print(fmt.Sprintf("proof %s | msgHeader %s | requester %s| ", proofHeader, msgHeader, Requester))
	if !strings.HasPrefix(auth, "Basic ") {
		log.Print("Invalid authorization:", auth)
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm=%q`, "Wolk"))
		http.Error(w, "BAD", 401)
		//http.Error(w, http.StatusText(unauth), unauth)
		return
	}
	log.Print("IS BASIC AUTH")
	up, err := base64.StdEncoding.DecodeString(auth[6:])
	if err != nil {
		log.Print("authorization decode error:", err)
		http.Error(w, http.StatusText(unauth), unauth)
		return
	}
	if string(up) != userPass {
		log.Print("invalid username:password:", string(up))
		http.Error(w, http.StatusText(unauth), unauth)
		return
	}
	log.Print("ISVALID")
	io.WriteString(w, "Goodbye, World!")

	/*
	   	// get jwk public key
	   	// get signature

	   	jwk := `{"crv":"P-256","ext":true,"key_ops":["verify"],"kty":"EC","x":"lxYtM63Yccu7xtRoZB9lTxbpdGjvJ2mYsZ4XmYUJ2WE","y":"zEBdoEflfZ3QNe_fGIjt3nyQ2clqBXimbaHHr6X1ik8"}`
	   	hashbytes := common.FromHex("e7e8b89c2721d290cc5f55425491ecd6831355e91063f20b39c22f9ec6a71f91")
	   	sig := common.FromHex("635ca4f4dcc66449804f7a255ee00ace03e0a650bd40eac85bf3f403693297a94d5eb308cad8be621e1d3c7d47d307f06a7a2f956c9e88c579527d9af7782c0d")

	           verified, err := JWKVerifyECDSA([]byte(message), jwkPublicKeyString, signature)
	           if err != nil {
	   	     http.Error("JWKVerifyECDSA %v\n", err)
	           }
	   // if not verified
	*/
	fmt.Fprintf(w, fmt.Sprintf("Dare to shine!"))
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm=%q`, "WOLK"))

	/*
	   # go test -run TestECDSA
	   Public Key: {"kty":"EC","crv":"P-256","x":"_xR3kdqo9DjvrfAzkSSszl9eIX3rdcFqGisSxVrcdMs","y":"jsrrqkWUCM5HMR4-mbyaXI28I09CSGv2-Z9zX3AIvUQ"}
	   Address: 3355b8aa79d21ae50e71495d172ff0df73d2c128
	   Signature: 30440220b91968b9b5f301a8110d334dc732b150bb26e7b4b5180088879852d43b6f605e02209df058273e9040a40f50249f84edcc30153e5bc293f8b1021e89ed581dd22d82
	   Verified: true len(signature)=70
	   PASS
	   ok	github.com/wolkdb/cloudstore/crypto	0.036s
	*/
	message := "test message"
	// get jwk public key
	jwkPublicKeyString := `{"kty":"EC","crv":"P-256","x":"_xR3kdqo9DjvrfAzkSSszl9eIX3rdcFqGisSxVrcdMs","y":"jsrrqkWUCM5HMR4-mbyaXI28I09CSGv2-Z9zX3AIvUQ"}`

	// get signature
	signature := common.FromHex("30440220b91968b9b5f301a8110d334dc732b150bb26e7b4b5180088879852d43b6f605e02209df058273e9040a40f50249f84edcc30153e5bc293f8b1021e89ed581dd22d82")

	verified, err := crypto.JWKVerifyECDSA([]byte(message), jwkPublicKeyString, signature)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	if verified {
		fmt.Fprintf(w, fmt.Sprintf("Not today, death!"))
	} else {
		http.Error(w, err.Error(), 401)
		fmt.Fprintf(w, fmt.Sprintf("All men must die!"))
		return
	}
}

func (s *MiniServer) Start() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.HTTPPort),
		Handler:      s.Handler,
		ReadTimeout:  600 * time.Second,
		WriteTimeout: 600 * time.Second,
	}
	SSLCertFile := "/etc/pki/tls/certs/wildcard/www.wolk.com.crt"
	SSLKeyFile := "/etc/pki/tls/certs/wildcard/www.wolk.com.key"

	err := srv.ListenAndServeTLS(SSLCertFile, SSLKeyFile)
	if err != nil {
		fmt.Printf("[http:Start] ListenAndServeTLS error %v", err)
	}
	fmt.Printf("[http:Start] ListenAndServeTLS %d", s.HTTPPort)
}

const userPass = "rosetta:code"
const unauth = http.StatusUnauthorized

func hw(w http.ResponseWriter, req *http.Request) {
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Basic ") {
		log.Print("Invalid authorization:", auth)
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm=%q`, "Wolk"))
		http.Error(w, http.StatusText(unauth), unauth)
		return
	}
	log.Print("try")
	up, err := base64.StdEncoding.DecodeString(auth[6:])
	if err != nil {
		log.Print("authorization decode error:", err)
		http.Error(w, http.StatusText(unauth), unauth)
		return
	}
	log.Print("try2")
	if string(up) != userPass {
		log.Print("invalid username:password:", string(up))
		http.Error(w, http.StatusText(unauth), unauth)
		return
	}
	io.WriteString(w, "Goodbye, World!")
}

func Test401(t *testing.T) {
	SSLCertFile := "/etc/pki/tls/certs/wildcard/www.wolk.com.crt"
	SSLKeyFile := "/etc/pki/tls/certs/wildcard/www.wolk.com.key"
	http.HandleFunc("/", hw)
	log.Fatal(http.ListenAndServeTLS(":8080", SSLCertFile, SSLKeyFile, nil))
}

func TestListenAndServeAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	s := &MiniServer{
		HTTPPort: 8080,
	}
	mux := http.NewServeMux()
	//mux.HandleFunc("/", hw)
	mux.HandleFunc("/", s.getConnection)
	s.Handler = mux
	s.Start()
}

func TestSSLCheck(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	for i := 0; i < 17; i++ {
		base := "cloudflare"
		if i < 8 {
			base = fmt.Sprintf("c%d", i)
		} else if i < 16 {
			base = fmt.Sprintf("s%d", i-8)
		}
		url := fmt.Sprintf("https://%s.wolk.com/healthcheck", base)
		req, err := http.NewRequest(http.MethodGet, url, nil)

		httpclient := http.Client{Timeout: time.Second * 5, Transport: DefaultTransport}
		resp, err := httpclient.Do(req)
		if err != nil {
			fmt.Printf("%s Do ERR %v\n", url, err)
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("%s ReadAll %v\n", url, err)
			} else {
				resp.Body.Close()
				fmt.Printf("SUCC %s %s\n", url, string(body))
			}
		}
	}
}

/*
// Content Header Inference
# go test -run Filetype
File SampleXLSFile_19kb.xls => extension: [doc] MIME:[application/msword]
File SampleVideo_1280x720_1mb.mp4 => extension: [mp4] MIME:[video/mp4]
File SampleVideo_1280x720_1mb.flv => extension: [flv] MIME:[video/x-flv]
File SamplePPTFile_500kb.ppt => extension: [doc] MIME:[application/msword]
File SamplePDFFile_5mb.pdf => extension: [pdf] MIME:[application/pdf]
File SampleCSVFile_11kb.csv => extension: [unknown] MIME:[]
File SampleAudio_0.4mb.mp3 => extension: [mp3] MIME:[audio/mpeg]
File banana.gif => extension: [gif] MIME:[image/gif]
File test.json => extension: [unknown] MIME:[]
File test.html => extension: [unknown] MIME:[]
File screenshot.png => extension: [png] MIME:[image/png]
File zone.txt => extension: [unknown] MIME:[]
File pot.jpeg => extension: [jpg] MIME:[image/jpeg]
File cat.jpeg => extension: [jpg] MIME:[image/jpeg]
PASS
ok	github.com/wolkdb/cloudstore/wolk	0.050s
*/

func TestFiletype(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	approved := []string{
		"SampleXLSFile_19kb.xls",
		"SampleVideo_1280x720_1mb.mp4",
		"SampleVideo_1280x720_1mb.flv",
		"SamplePPTFile_500kb.ppt",
		"SamplePDFFile_5mb.pdf",
		"SampleCSVFile_11kb.csv",
		"SampleAudio_0.4mb.mp3",
		"banana.gif",
		"test.json",
		"test.html",
		"screenshot.png",
		"zone.txt",
		"pot.jpeg",
		"cat.jpeg",
	}

	dir := "/root/go/src/github.com/wolkdb/cloudstore/content"
	for _, fn := range approved {
		fullpath := path.Join(dir, fn)
		if _, err := os.Stat(fullpath); os.IsNotExist(err) {
			t.Fatalf("Stat %v\n", err)
		}

		buf, _ := ioutil.ReadFile(fullpath)
		kind, unknown := filetype.Match(buf)
		if unknown != nil {
			t.Fatalf("Unknown %s\n", unknown)
		}
		fmt.Printf("http://cloudflare.wolk.com/file?fn=%s\t%s\n", fn, kind.MIME.Value)
	}
}
