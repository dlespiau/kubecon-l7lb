package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var service struct {
	duration Rander
}

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
	time.Sleep(time.Duration(service.duration.Rand() * float64(time.Millisecond)))
}

func main() {
	listen := flag.String("listen", ":8080", "address the service will listen on")
	durationModel := flag.String("duration.model", "log-normal", "enable request duration modelling")
	constant := flag.Float64("duration.constant.ms", 10, "time (in ms) of request duration")
	mu := flag.Float64("duration.log-normal.mu", 3, "μ parameter of the lognormal distribution of request duration")
	sigma := flag.Float64("duration.log-normal.sigma", 0.4, "σ parameter of the lognormal distribution of request duration")
	flag.Parse()

	switch *durationModel {
	case "zero":
		service.duration = &Constant{0}
	case "constant":
		service.duration = &Constant{*constant}
	case "log-normal":
		service.duration = &LogNormal{Mu: *mu, Sigma: *sigma}
	default:
		log.Fatalf("unknown duration model: %s", *durationModel)
	}

	http.HandleFunc("/ready", ready)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/hostname", hostname)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
