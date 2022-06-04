package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/filecoin-project/saturn-l2/resources"
)

type config struct {
	FilAddr string `json:"fil_wallet_address"`
}

func main() {
	var port int
	portStr := os.Getenv("PORT")
	if portStr == "" {
		port = 5500
	} else {
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			panic(fmt.Errorf("Invalid PORT value '%s': %s", portStr, err.Error()))
		}
	}

	filAddr := os.Getenv("FIL_WALLET_ADDRESS")
	if filAddr == "" {
		panic(fmt.Errorf("no FIL_WALLET_ADDRESS provided"))
	}
	conf, err := json.Marshal(config{FilAddr: filAddr})
	if err != nil {
		panic(errors.New("failed to serialize config"))
	}

	m := mux.NewRouter()
	m.PathPrefix("/config").Handler(http.HandlerFunc(configHandler(conf)))
	m.PathPrefix("/webui").Handler(http.HandlerFunc(webuiHandler))

	srv := &http.Server{
		Handler: m,
	}

	nl, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	go func() {
		if err := srv.Serve(nl); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	port = nl.Addr().(*net.TCPAddr).Port
	fmt.Println("Server listening on", nl.Addr())
	fmt.Printf("WebUI: http://localhost:%d/webui\n", port)
	for {

	}
}

func webuiHandler(w http.ResponseWriter, r *http.Request) {
	rootDir := "webui"
	path := strings.TrimPrefix(r.URL.Path, "/")

	_, err := resources.WebUI.Open(path)
	if path == rootDir || os.IsNotExist(err) {
		// file does not exist, serve index.html
		index, err := resources.WebUI.ReadFile(rootDir + "/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(index)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.FS(resources.WebUI)).ServeHTTP(w, r)
}

func configHandler(conf []byte) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(conf)
	}
}
