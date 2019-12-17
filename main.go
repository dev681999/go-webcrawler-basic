package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"

	"mvdan.cc/xurls/v2"
)

// Fetcher is an interface to fetcch urls from a given url
type Fetcher interface {
	Fetch(url string) (urls []string, err error)
}

var fetched = struct {
	m map[string]error
	sync.Mutex
}{m: make(map[string]error)}

var errLoading = errors.New("url load in progress")

// Crawl uses fetcher to recursively crawl pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	fmt.Printf("Found: %v\n", url)
	if depth <= 0 {
		return
	}

	fetched.Lock()
	if _, ok := fetched.m[url]; ok {
		fetched.Unlock()
		return
	}

	fetched.m[url] = errLoading
	fetched.Unlock()

	urls, err := fetcher.Fetch(url)

	fetched.Lock()
	fetched.m[url] = err
	fetched.Unlock()

	if err != nil {
		return
	}

	done := make(chan bool)
	for _, u := range urls {
		go func(url string) {
			Crawl(url, depth-1, fetcher)
			done <- true
		}(u)
	}
	for range urls {
		<-done
	}
}

func main() {
	Crawl("https://google.com/", 2, fetcher)
}

type httpFetcher struct {
	re *regexp.Regexp
}

func (f *httpFetcher) Fetch(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	urls := f.re.FindAllString(string(body), -1)
	return urls, nil
}

var fetcher = &httpFetcher{
	re: xurls.Relaxed(),
}
