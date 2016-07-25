package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	layoutNotFound            = "404.html"
	layoutBadRequest          = "400.html"
	layoutInternalServerError = "500.html"
)

func render(context *gin.Context, layout string, status ...int) {
	if len(status) == 0 {
		status = []int{http.StatusOK}
	} else if len(status) > 1 {
		panic("too much statuses")
	}

	context.Set("render", true)
	context.HTML(status[0], layout, context.Keys)
}

func renderBadRequest(context *gin.Context) {
	render(context, layoutBadRequest, http.StatusBadRequest)
}

func renderNotFound(context *gin.Context) {
	render(context, layoutNotFound, http.StatusNotFound)
}

func renderInternalServerError(context *gin.Context) {
	render(context, layoutInternalServerError, http.StatusInternalServerError)
}
