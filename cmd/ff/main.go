package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/nakatanakatana/ff"
)

const (
	HTTPReadTimeout  = 30 * time.Second
	HTTPWriteTimeout = 30 * time.Second
)

var latestOnlyFlag bool

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

func main() {
	muteAuthors := strings.Split(os.Getenv("MUTE_AUTHORS"), ",")
	muteURLs := strings.Split(os.Getenv("MUTE_URLS"), ",")

	latestOnly := os.Getenv("LATEST_ONLY")
	if latestOnly != "" {
		latestOnlyFlag = true
	}

	filtersMap := ff.CreateFiltersMap(muteAuthors, muteURLs)
	modifiersMap := ff.CreateModifierMap()
	mux := http.NewServeMux()
	mux.Handle("/", createHandler(filtersMap, modifiersMap))

	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  HTTPReadTimeout,
		WriteTimeout: HTTPWriteTimeout,
	}

	log.Fatal(server.ListenAndServe())
}
