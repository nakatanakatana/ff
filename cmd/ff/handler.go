package main

import (
	"fmt"
	"net/http"

	"github.com/mmcdole/gofeed"
	"github.com/nakatanakatana/ff"
)

func createHandler(filtersMap ff.FilterFuncMap, modifiersMap ff.ModifierFuncMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			fmt.Fprintln(w, fmt.Errorf("ParseURL Error: %w", err))

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
}
