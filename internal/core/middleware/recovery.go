package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/Nzyazin/itk/internal/core/logger"
)

func Recovery(log logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if rec := recover(); rec != nil {
                    log.Error("panic recovered",
                        logger.StringField("path", r.URL.Path),
                        logger.AnyField("error", rec),
                        logger.StringField("stack", string(debug.Stack())),
                    )
                    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}