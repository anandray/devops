package main

import (
       "log"
	"net/http"

)

func HelloServer(w http.ResponseWriter, req *http.Request) {
    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("This is an example server.\n"))
}

func main() {
    http.HandleFunc("/", HelloServer)
	SSLCertFile := "/etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt"
	SSLKeyFile := "/etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key"
    err := http.ListenAndServeTLS(":88", SSLCertFile, SSLKeyFile, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

