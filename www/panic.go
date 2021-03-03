package www

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func handlePanic(ctx *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			var brokenPipe bool
			if ne, ok := err.(*net.OpError); ok {
				if se, ok := ne.Err.(*os.SyscallError); ok {
					if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
						brokenPipe = true
					}
				}
			}

			if brokenPipe {
				ctx.Error(err.(error))
				ctx.Abort()
			} else {
				fmt.Printf("%+v", errors.WithStack(err.(error)))
				sentry.CaptureException(err.(error))

				ctx.Status(http.StatusInternalServerError)
				ctx.Abort()
			}
		}
	}()
	ctx.Next()
}
