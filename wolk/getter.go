package wolk

import (
	//        wolkcommon "github.com/wolkdb/cloudstore/common"

	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	//        "net"
	"net/http"
	"time"
)

type ClientGetter struct {
	url string
}

func (g *ClientGetter) Get(ctx context.Context, key Reference) (ChunkData, error) {
	url := fmt.Sprintf("%s/wolk/chunk/%x", g.url, key)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	httpclient := &http.Client{Timeout: time.Second * 1000, Transport: DefaultTransport}

	res, err := httpclient.Do(req)
	if err != nil {
		fmt.Printf("ERR: %v\n", err)
	} else {
		// fmt.Printf("%s %s (%x) %v\n", run, url, k, res.StatusCode)
		defer res.Body.Close()
		if res.StatusCode == 200 {
			body, reserr := ioutil.ReadAll(res.Body)
			var c SetShareResponse
			err = json.Unmarshal(body, &c)
			if reserr != nil {
				fmt.Printf("DONE %d %s\n", res.StatusCode, body)
			} else {
				// fmt.Printf("%s", body)
				return body, nil
			}
		}
	}
	return nil, err

}

func NewGetter(url string) *ClientGetter {
	return &ClientGetter{url: url}
}
