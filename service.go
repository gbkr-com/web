package web

import (
	"context"
	"log"
	"net/http"
)

// Request for service.
type Request struct {
	method string
	uri    string
	body   string
	reply  chan *Response
}

// MakeRequest returns a new service request, ready to send.
func MakeRequest(method, URI, body string) *Request {
	return &Request{
		method: method,
		uri:    URI,
		body:   body,
		reply:  make(chan *Response, 1),
	}
}

// Response is the result of serving a request.
type Response struct {
	Status int
	Body   interface{}
}

// RespondWith returns the response.
func (r *Request) RespondWith(status int, body interface{}) {
	r.reply <- &Response{Status: status, Body: body}
	close(r.reply)
}

// Serve requests.
func Serve(ctx context.Context, requests chan *Request, logger *log.Logger) {
	//
	// The initial greeting.
	//
	greeting := "Hello World"
	//
	// The main event loop.
	//
	for {
		select {
		//
		// The application may be terminated.
		//
		case <-ctx.Done():
			logger.Print("serve terminating")
			return
		//
		// Otherwise handle requests.
		//
		case req := <-requests:
			if req == nil {
				continue
			}
			switch req.method {

			case http.MethodGet:
				var body struct {
					Greeting string
				}
				body.Greeting = greeting
				req.RespondWith(http.StatusOK, body)

			case http.MethodPut:
				greeting = req.body
				req.RespondWith(http.StatusNoContent, nil)

			default:
				req.RespondWith(http.StatusBadRequest, nil)

			}

		}
	}
}
