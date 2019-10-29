import (
	"net/http"
	"fmt"
	"io/ioutil"
	"ar"
        "github.com/facebookgo/grace/gracehttp"
)

func messageHandler(rw http.ResponseWriter, r *http.Request) {
	msg, err := ioutil.ReadAll(r.Body)
 	if err != nil {
 		fmt.Println(err)
 		// os.Exit(1)
 	}

	fmt.Println(string(msg))
	resp := ar.Process_message(string(msg))
	fmt.Fprintf(rw, resp)
}

func main() {
	http.HandleFunc("/", messageHandler)
	fmt.Println("Listening on 9001")
	http.ListenAndServe(":9001", nil)

}
