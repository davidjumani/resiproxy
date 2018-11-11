package rest

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/marqub/resiproxy/k8s"
	"github.com/marqub/resiproxy/log"

	"github.com/sirupsen/logrus"
)

type Proxy struct {
	Name     string   `json:"name"`
	Listen   string   `json:"name"`
	Upstream string   `json:"name"`
	Enabled  string   `json:"name"`
	Toxics   []string `json:"name"`
}

func CreateProxy(w http.ResponseWriter, r *http.Request) {

	var proxy Proxy
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	if err := json.Unmarshal(body, &proxy); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(jsonErr{Code: http.StatusPreconditionFailed, Text: "Invalid Proxy"}); err != nil {
			panic(err)
		}
		return
	}
	log.Logger().Info("Create proxy :", string(body))
	err = k8s.CreateK8sMapping(proxy.Listen)
	if err != nil {
		logrus.Error("K8s mappings can not be created")
		return
	}
	ProxyRequest(w, r)
}

func ProxyRequest(w http.ResponseWriter, r *http.Request) {
	log.Logger().Info("Proxy request")
	serveReverseProxy(getEnv("key", "http://www.google.com"), w, r)
}

// Get env var or default
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	// parse the url
	url, _ := url.Parse(target)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}