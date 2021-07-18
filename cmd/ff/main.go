package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mmcdole/gofeed"
	"github.com/nakatanakatana/ff"
)

var latestOnlyFlag bool

func init() {
	latestOnly := os.Getenv("LATEST_ONLY")
	if latestOnly != "" {
		latestOnlyFlag = true
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	upstream, ok := queries["url"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "must set URL")
		return
	}
	if len(upstream) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "cannot set multiple URL")
		return
	}
	u := upstream[0]
	fp := gofeed.NewParser()
	originFeed, err := fp.ParseURL(u)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err)
		return
	}
	var filters []ff.FilterFunc
	if latestOnlyFlag {
		filters = append(filters, ff.CreateFilter("latest", ""))
	}
	for key, values := range queries {
		for _, v := range values {
			f := ff.CreateFilter(key, v)
			if f != nil {
				filters = append(filters, f)
			}
		}
	}
	filteredFeed, err := ff.Filter(originFeed, filters...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}
	c := ff.Convert(filteredFeed)
	rss, err := c.ToRss()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, rss)
	return
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handler))
	log.Fatal(http.ListenAndServe(":8080", mux))
}
