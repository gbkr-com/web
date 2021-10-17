package web

import (
	"context"
	"net/http"

	"github.com/gbkr-com/sharding"
)

// The combined references to the response writer and the request, to be written
// to a channel.
//
type data struct {
	writer  http.ResponseWriter
	request *http.Request
}

// The archetypical handling of a channel.
//
func archetype(handler http.HandlerFunc, channel chan data) {
	for d := range channel {
		handler(d.writer, d.request)
	}
}

// Closer allows pooled handlers to be closed in a regulated way.
//
type Closer interface {

	// Close using the given context, allowing closing to terminate after a
	// deadline or timeout. If the context is nil, the behaviour for
	// context.Background() is assuemd.
	//
	Close(ctx context.Context)
}

type pool struct {
	channel chan data
}

// Close implements Closer.
//
func (p *pool) Close(ctx context.Context) {
	close(p.channel)
	if ctx != nil {
		<-ctx.Done()
	}
}

// NewPooledHandler returns an http.HandlerFunc which writes requests to any of
// N goroutines running the given handler. N must be greater than one.
//
func NewPooledHandler(n int, handler http.HandlerFunc) (http.HandlerFunc, Closer, error) {
	if n < 2 {
		return nil, nil, ErrBadPool
	}
	control := &pool{channel: make(chan data, n)}
	for i := 0; i < n; i++ {
		go archetype(handler, control.channel)
	}
	//
	// The returned handler writes to the single channel for any of the
	// goroutines to process.
	//
	h := func(writer http.ResponseWriter, request *http.Request) {
		control.channel <- data{writer, request}
	}
	return h, control, nil
}

type sharded struct {
	scheme   *sharding.Scheme
	channels []chan data
}

// Close implements Closer.
//
func (s *sharded) Close(ctx context.Context) {
	for _, ch := range s.channels {
		close(ch)
	}
	if ctx != nil {
		<-ctx.Done()
	}
}

// ShardKeyFunction extracts the shard key, as a byte slice, from a request.
//
type ShardKeyFunction func(*http.Request) []byte

// NewShardedHandler returns an http.HandlerFunc which writes requests to N
// goroutines running the given handler, with each request distributed by a
// sharding scheme. N must be a power of two greater than one.
//
// M defines the length of each channel into the goroutines. It must be greater
// than zero to avoid blocking.
//
func NewShardedHandler(n, m int, handler http.HandlerFunc, f ShardKeyFunction) (http.HandlerFunc, Closer, error) {
	sh, err := sharding.New(uint64(n))
	if err != nil {
		return nil, nil, err
	}
	if m < 1 {
		return nil, nil, ErrBadQueue
	}
	control := &sharded{
		scheme:   sh,
		channels: make([]chan data, n),
	}
	for i := range control.channels {
		control.channels[i] = make(chan data, m)
		go archetype(handler, control.channels[i])
	}
	//
	// The returned handler extracts the shard key from the request and writes
	// to the appropriate channel.
	//
	h := func(writer http.ResponseWriter, request *http.Request) {
		i := control.scheme.WithBytes(f(request))
		control.channels[i] <- data{writer, request}
	}
	return h, control, nil
}
