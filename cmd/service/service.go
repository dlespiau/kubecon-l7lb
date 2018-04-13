package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func ready(w http.ResponseWriter, r *http.Request) {
}

func echo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("invalid method: %s", r.Method), http.StatusBadRequest)
		return
	}
	io.Copy(w, r.Body)
}

func hostname(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadFile("/etc/hostname")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(body))
}

func main() {
	listen := flag.String("listen", ":8080", "address the service will listen on")
	flag.Parse()

	http.HandleFunc("/ready", ready)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/hostname", hostname)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
