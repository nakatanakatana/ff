package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/nakatanakatana/ff"
)

const (
	HTTPReadTimeout  = 30 * time.Second
	HTTPWriteTimeout = 30 * time.Second
)

var (
	latestOnlyFlag bool
	filtersMap     ff.FilterFuncMap
	modifiersMap   ff.ModifierFuncMap
)

func parseQueries(queries url.Values,
	filtersMap ff.FilterFuncMap,
	modifiersMap ff.ModifierFuncMap) ([]ff.FilterFunc,
	[]ff.ModifierFunc,
) {
	var filters []ff.FilterFunc
	if latestOnlyFlag {
		filters = append(filters, ff.CreateFilter("published_at.latest", "", filtersMap))
		filters = append(filters, ff.CreateFilter("updated_at.latest", "", filtersMap))
	}

	f, m := ff.ParseQueries(queries, filtersMap, modifiersMap)
	filters = append(filters, f...)

	return filters, m
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

	filters, modifiers := parseQueries(queries, filtersMap, modifiersMap)

	filteredFeed, err := ff.Apply(originFeed, filters, modifiers)
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
	modifiersMap = ff.CreateModifierMap()
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handler))

	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  HTTPReadTimeout,
		WriteTimeout: HTTPWriteTimeout,
	}

	log.Fatal(server.ListenAndServe())
}
