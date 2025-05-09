package middleware

import (
	"net/http"

	"github.com/Nzyazin/itk/internal/core/logger"
)

type ErrorHandler struct {
    handler http.Handler
    log     logger.Logger
}

func WithErrorHandler(log logger.Logger) func(http.Handler) http.Handler {
    return func(h http.Handler) http.Handler {
        return &ErrorHandler{handler: h, log: log}
    }
}

func (eh *ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    defer func() {
        if err := recover(); err != nil {
            eh.log.Error("request processing failed",
                logger.StringField("method", r.Method),
                logger.StringField("path", r.URL.Path),
                logger.AnyField("error", err),
            )
            w.WriteHeader(http.StatusInternalServerError)
        }
    }()
    
    eh.handler.ServeHTTP(w, r)
}
