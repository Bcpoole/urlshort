package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bcpoole/urlshort"
)

func main() {
	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	// Build the YAMLHandler using the mapHandler as the fallback
	var yamlFile = flag.String("yamlfile", "urlmappings.yaml", "Provide absolute path for yaml file with redirect urls.")
	var jsonFile = flag.String("jsonfile", "urlmappings.json", "Provide absolute path for json file with redirect urls.")
	var boltFile = flag.String("boltfile", "bolt.db", "Provide absolute path for bolt db file with redirect urls.")
	flag.Parse()

	boltHandler, err := urlshort.BoltHandler(*boltFile, mapHandler)
	if err != nil {
		panic(err)
	}

	yaml, err := ioutil.ReadFile(*yamlFile)
	if err != nil {
		panic(err)
	}
	yamlHandler, err := urlshort.YAMLHandler(yaml, boltHandler)
	if err != nil {
		panic(err)
	}

	jsonData, err := ioutil.ReadFile(*jsonFile)
	jsonHandler, err := urlshort.JSONHandler(jsonData, yamlHandler)
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", jsonHandler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
