package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gbkr-com/app"
	"github.com/gbkr-com/web"
	"github.com/gin-gonic/gin"
)

const (
	envADDRESS  = "ADDRESS"
	envCAPACITY = "CAPACITY"
	envTIMEOUT  = "TIMEOUT"
)

const (
	defaultADDRESS  = ":8080"
	defaultCAPACITY = 256
	defaultTIMEOUT  = time.Second
)

func main() {
	logger := log.New(os.Stdout, "", log.Lmicroseconds)
	ctx, cxl, _ := app.WithCancel(context.Background())
	capacity := configureCapacity()
	requests := make(chan *web.Request, capacity)
	logger.Print("capacity ", capacity)
	//
	// Launch the service.
	//
	go func() {
		logger.Print("ready to serve")
		web.Serve(ctx, requests, logger)
	}()
	//
	// Launch the HTTP server.
	//
	h := &web.HandlerContext{
		Context:  ctx,
		Cancel:   cxl,
		Requests: requests,
		Logger:   logger,
	}
	g := gin.New()
	h.BindTo(g)
	addr := configureAddress()
	srv := &http.Server{
		Addr:    addr,
		Handler: g,
	}
	go func() {
		logger.Print("listening on ", addr)
		srv.ListenAndServe()
	}()
	//
	// Wait until SIGINT then gracefully shut down.
	//
	<-ctx.Done()
	logger.Print()
	logger.Print("shutting down")
	timeout := configureTimeout()
	shutdown(srv, timeout)
	logger.Print("done")
}

func configureCapacity() (capacity int) {
	capacity = defaultCAPACITY
	envCap, ok := os.LookupEnv(envCAPACITY)
	if !ok {
		return
	}
	capacity, err := strconv.Atoi(envCap)
	if err != nil {
		capacity = defaultCAPACITY
	}
	return
}

func configureAddress() (address string) {
	address = defaultADDRESS
	address, ok := os.LookupEnv(envADDRESS)
	if !ok {
		address = defaultADDRESS
	}
	return
}

func configureTimeout() (timeout time.Duration) {
	timeout = defaultTIMEOUT
	str, ok := os.LookupEnv(envTIMEOUT)
	if !ok {
		return
	}
	timeout, _ = time.ParseDuration(str)
	return
}

func shutdown(srv *http.Server, timeout time.Duration) {
	ctx, cxl := context.WithTimeout(context.Background(), timeout)
	defer cxl()
	srv.Shutdown(ctx)
}
