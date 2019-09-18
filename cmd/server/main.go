package main

import (
	"net"
	"net/http"
	"os"

	"encoding/json"
	"github.com/gorilla/mux"
)

var (
	hostname string
	ips      []string
)

func init() {
	hostname, _ = os.Hostname()
	addrs, _ := net.InterfaceAddrs()
	ips = make([]string, len(addrs))
	for index, addr := range addrs {
		ips[index] = addr.String()
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handle).Methods("GET")
	err := http.ListenAndServe("0.0.0.0:8080", r)
	if err != nil {
		panic(err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	_ = encoder.Encode(map[string]interface{}{
		"env":      os.Getenv("ENV"),
		"hostname": hostname,
		"ips":      ips,
	})
}
