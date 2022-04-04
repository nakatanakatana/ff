package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mmcdole/gofeed"
	"github.com/nakatanakatana/ff"
)

var (
	latestOnlyFlag bool
	filtersMap     ff.FilterFuncMap
)

func parseQueries(queries url.Values, filtersMap ff.FilterFuncMap) []ff.FilterFunc {
	var filters []ff.FilterFunc
	if latestOnlyFlag {
		filters = append(filters, ff.CreateFilter("latest", "", filtersMap))
	}

	for key, values := range queries {
		for _, v := range values {
			f := ff.CreateFilter(key, v, filtersMap)
			if f != nil {
				filters = append(filters, f)
			}
		}
	}

	return filters
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

	filters := parseQueries(queries, filtersMap)

	filteredFeed, err := ff.Apply(originFeed, filters...)
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
}

func main() {
	muteAuthors := strings.Split(os.Getenv("MUTE_AUTHORS"), ",")
	muteURLs := strings.Split(os.Getenv("MUTE_URLS"), ",")

	latestOnly := os.Getenv("LATEST_ONLY")
	if latestOnly != "" {
		latestOnlyFlag = true
	}

	filtersMap = ff.CreateFiltersMap(muteAuthors, muteURLs)
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handler))
	log.Fatal(http.ListenAndServe(":8080", mux))
}
