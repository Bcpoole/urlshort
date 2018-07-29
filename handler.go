package urlshort

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	yaml "gopkg.in/yaml.v2"
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path, ok := pathsToUrls[r.URL.Path]
		if ok {
			http.Redirect(w, r, path, http.StatusFound)
		} else {
			fallback.ServeHTTP(w, r)
		}
	}
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	ymlPaths := []map[string]string{}
	err := yaml.Unmarshal(yml, &ymlPaths)
	if err != nil {
		return nil, err
	}
	paths := buildRedirectMap(ymlPaths)

	return func(w http.ResponseWriter, r *http.Request) {
		path, ok := paths[r.URL.Path]
		if ok {
			http.Redirect(w, r, path, http.StatusFound)
		} else {
			fallback.ServeHTTP(w, r)
		}
	}, nil
}

// JSONHandler parses json []byte of url handler mappings an redirects base on those inputs.
// Else falls back to provided Handler.
func JSONHandler(data []byte, fallback http.Handler) (http.HandlerFunc, error) {
	jsonPaths := []map[string]string{}
	err := json.Unmarshal(data, &jsonPaths)
	if err != nil {
		return nil, err
	}
	paths := buildRedirectMap(jsonPaths)

	return func(w http.ResponseWriter, r *http.Request) {
		path, ok := paths[r.URL.Path]
		if ok {
			http.Redirect(w, r, path, http.StatusFound)
		} else {
			fallback.ServeHTTP(w, r)
		}
	}, nil
}

// BoltHandler reads a BoltDB of url handler mappings an redirects base on those inputs.
// Else falls back to provided Handler.
func BoltHandler(boltFile string, fallback http.Handler) (http.HandlerFunc, error) {
	db, err := bolt.Open(boltFile, 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// This bit of code is to be run if the Bolt file does not exist.
	db.Update(func(tx *bolt.Tx) error {
		b, err2 := tx.CreateBucket([]byte("URLRedirects"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}
		err := b.Put([]byte("/urlshort-bolt"), []byte("https://github.com/bcpoole/urlshort"))
		if err != nil {
			return fmt.Errorf("put: %s", err2)
		}
		return nil
	})

	paths := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("URLRedirects"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			paths[string(k)] = string(v)
		}

		return nil
	})

	return func(w http.ResponseWriter, r *http.Request) {
		path, ok := paths[r.URL.Path]
		if ok {
			http.Redirect(w, r, path, http.StatusFound)
		} else {
			fallback.ServeHTTP(w, r)
		}
	}, nil
}

func buildRedirectMap(data []map[string]string) map[string]string {
	redirects := make(map[string]string)
	for _, m := range data {
		redirects[m["path"]] = m["url"]
	}
	return redirects
}
