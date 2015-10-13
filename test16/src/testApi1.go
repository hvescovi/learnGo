package main

import (
  "net/http"
  "fmt"
  //"os"
  "io/ioutil"
  //"strings"
  "net/url"
  //"testing"
  "reflect"
  "strconv"
  "github.com/coreos/etcd/pkg/pathutil"
)


func main() {
	ep := url.URL{Scheme: "http", Host: "127.0.0.1:4001", Path: "/v2/keys"}
	baseWantURL := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:4001",
		Path:   "/v2/keys/foo",
	}
	wantHeader := http.Header{}

  f := getAction{
			Key:       "/foo",
			Recursive: false,
			Sorted:    false,
			Quorum:    false,
	}

	got := *f.HTTPRequest(ep)

	wantURL := baseWantURL
	wantURL.RawQuery = "quorum=false&recursive=false&sorte=false"

  //gotBytes, err := ioutil.ReadAll(got.Body)

	err := assertRequest(got, "GET", wantURL, wantHeader, nil)
	if err != nil {
		fmt.Println(err) //t.Errorf("#%d: %v", i, err)
	}
}

func assertRequest(got http.Request, wantMethod string, wantURL *url.URL, wantHeader http.Header, wantBody []byte) error {
	if wantMethod != got.Method {
		return fmt.Errorf("want.Method=%#v got.Method=%#v", wantMethod, got.Method)
	}

	if !reflect.DeepEqual(wantURL, got.URL) {
		return fmt.Errorf("want.URL=%#v got.URL=%#v", wantURL, got.URL)
	}

	if !reflect.DeepEqual(wantHeader, got.Header) {
		return fmt.Errorf("want.Header=%#v got.Header=%#v", wantHeader, got.Header)
	}

	if got.Body == nil {
		if wantBody != nil {
			return fmt.Errorf("want.Body=%v got.Body=%v", wantBody, got.Body)
		}
	} else {
		if wantBody == nil {
			return fmt.Errorf("want.Body=%v got.Body=%s", wantBody, got.Body)
		} else {
			gotBytes, err := ioutil.ReadAll(got.Body)
			if err != nil {
				return err
			}

			if !reflect.DeepEqual(wantBody, gotBytes) {
				return fmt.Errorf("want.Body=%s got.Body=%s", wantBody, gotBytes)
			}
		}
	}

	return nil
}

type getAction struct {
	Prefix    string
	Key       string
	Recursive bool
	Sorted    bool
	Quorum    bool
}

func (g *getAction) HTTPRequest(ep url.URL) *http.Request {
	u := v2KeysURL(ep, g.Prefix, g.Key)

	params := u.Query()
	params.Set("recursive", strconv.FormatBool(g.Recursive))
	params.Set("sorted", strconv.FormatBool(g.Sorted))
	params.Set("quorum", strconv.FormatBool(g.Quorum))
	u.RawQuery = params.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	return req
}

func v2KeysURL(ep url.URL, prefix, key string) *url.URL {
	// We concatenate all parts together manually. We cannot use
	// path.Join because it does not reserve trailing slash.
	// We call CanonicalURLPath to further cleanup the path.
	if prefix != "" && prefix[0] != '/' {
		prefix = "/" + prefix
	}
	if key != "" && key[0] != '/' {
		key = "/" + key
	}
	ep.Path = pathutil.CanonicalURLPath(ep.Path + prefix + key)
	return &ep
}
