package web

import (
	"context"
	"log"
	"net/http"

	"github.com/gbkr-com/app"
	"github.com/gin-gonic/gin"
)

// HandlerContext is the context for all interactions. The Context is for the
// whole application, not each request. Likewise, if handlers need to, they can
// use the Cancel function to signal to other goroutines that the application
// must terminate.
type HandlerContext struct {
	Context  context.Context
	Cancel   context.CancelFunc
	Requests chan *Request
	Logger   *log.Logger
}

// BindTo binds the HTTP handler functions to gin.
func (c *HandlerContext) BindTo(router *gin.Engine) {
	grp := router.Group("/v1")
	grp.GET("/greet", c.Greet)
	grp.PUT("/greeting", c.Greeting)
}

// Greet is the HTTP handler for getting the current greeting.
func (c *HandlerContext) Greet(ctx *gin.Context) {
	//
	// If the app is cancelled do not proceed with this.
	//
	if app.IsDone(c.Context) {
		ctx.Status(http.StatusServiceUnavailable)
		return
	}
	req := MakeRequest(http.MethodGet, "", "")
	c.Requests <- req
	res := <-req.reply
	ctx.JSON(res.Status, res.Body)
}

// Greeting is the HTTP handler to change the greeting.
func (c *HandlerContext) Greeting(ctx *gin.Context) {
	//
	// If the app is cancelled do not proceed with this.
	//
	if app.IsDone(c.Context) {
		ctx.Status(http.StatusServiceUnavailable)
		return
	}
	//
	// Parse the HTTP body.
	//
	var body struct {
		Greeting string
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	req := MakeRequest(http.MethodPut, ctx.FullPath(), body.Greeting)
	c.Requests <- req
	res := <-req.reply
	ctx.Status(res.Status)
}
