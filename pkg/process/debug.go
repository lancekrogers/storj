// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package process

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"

	"go.uber.org/zap"
	monkit "gopkg.in/spacemonkeygo/monkit.v2"
	"gopkg.in/spacemonkeygo/monkit.v2/present"

	"storj.io/storj/internal/version"
)

var (
	debugAddr = flag.String("debug.addr", "127.0.0.1:0", "address to listen on for debug endpoints")
)

func init() {
	// zero out the http.DefaultServeMux net/http/pprof so unhelpfully
	// side-effected.
	*http.DefaultServeMux = http.ServeMux{}
}

func initDebug(logger *zap.Logger, r *monkit.Registry) (err error) {
	var mux http.ServeMux
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Handle("/version/", http.StripPrefix("/version", http.HandlerFunc(version.DebugHandler)))
	mux.Handle("/mon/", http.StripPrefix("/mon", present.HTTP(r)))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "OK")
	})
	ln, err := net.Listen("tcp", *debugAddr)
	if err != nil {
		return err
	}
	go func() {
		logger.Debug(fmt.Sprintf("debug server listening on %s", ln.Addr().String()))
		err := (&http.Server{Handler: &mux}).Serve(ln)
		if err != nil {
			logger.Error("debug server died", zap.Error(err))
		}
	}()
	return nil
}
