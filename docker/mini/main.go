
package main

import (
    "net/http"
)

type mrHandler struct {
}
type msHandler struct {
}

func (m *mrHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello mr 32007"))
}

func (m *msHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello ms 2087"))
}

func main() {

     go func() {
     	       http.ListenAndServe(":32007", &mrHandler{})
       }()

       http.ListenAndServe(":2087", &msHandler{})

}