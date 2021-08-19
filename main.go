package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"
)

const (
	port = "8443"
)

var (
	tlscert, tlskey string
)

func validateServe(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Called validate")
	serve(w, r, validate)
}

func mutateServe(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Called mutate")
	serve(w, r, mutate)
}

func main() {

	flag.StringVar(&tlscert, "tlsCertFile", "/etc/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&tlskey, "tlsKeyFile", "/etc/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")

	flag.Parse()

	certs, err := tls.LoadX509KeyPair(tlscert, tlskey)
	if err != nil {
		klog.Errorf("Failed to load key pair: %v", err)
	}

	server := &http.Server{
		Addr:      fmt.Sprintf(":%v", port),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{certs}},
	}

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", validateServe)
	mux.HandleFunc("/mutate", mutateServe)
	server.Handler = mux

	// start webhook server in new routine
	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			klog.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	klog.Infof("Server running & listening on port: %s", port)

	// listening to shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	klog.Info("Got shutdown signal, shutting down webhook server gracefully...")
	server.Shutdown(context.Background())
}
