package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
)

const name = "serve"

const version = "0.0.4"

var revision = "HEAD"

func main() {
	addr := flag.String("a", ":5000", "address to serve(host:port)")
	prefix := flag.String("p", "/", "prefix path under")
	root := flag.String("r", ".", "root path to serve")
	certFile := flag.String("cf", "", "tls cert file")
	keyFile := flag.String("kf", "", "tls key file")
	gziped := flag.Bool("gzip", false, "enable gzip compression")
	zstd := flag.Bool("zstd", false, "enable zstd compression")
	showVersion := flag.Bool("v", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s %s (rev: %s/%s)\n", name, version, revision, runtime.Version())
		return
	}

	var err error
	*root, err = filepath.Abs(*root)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("serving %s as %s on %s", *root, *prefix, *addr)

	http.Handle(*prefix, http.StripPrefix(*prefix, http.FileServer(http.Dir(*root))))

	var handler http.Handler = http.DefaultServeMux
	handler = LogMiddleware(handler)

	if *gziped {
		handler = GzipCompressMiddleware(handler)
	}

	if *zstd {
		handler = ZstdCompressMiddleware(handler)
	}

	if *certFile != "" && *keyFile != "" {
		err = http.ListenAndServeTLS(*addr, *certFile, *keyFile, handler)
	} else {
		err = http.ListenAndServe(*addr, handler)
	}
	if err != nil {
		log.Fatalln(err)
	}
}
