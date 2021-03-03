package www

import (
	"net/http"
	"path/filepath"

	"github.com/RyuaNerin/ShareMe/share"
	"github.com/gin-gonic/gin"
)

func initGin(g *gin.Engine) {
	g.Use(handlePanic)
	g.MaxMultipartMemory = 128 * 1024 // 128 KiB 이상은 모조리 임시파일로

	g.LoadHTMLGlob(filepath.Join(share.DirPublic, "*.htm"))

	g.Static("/static", filepath.Join(share.DirPublic, "static"))
	g.POST("/upload", handleUpload)
	g.GET("/download", handleDownload)

	g.StaticFile("/", filepath.Join(share.DirPublic, "index.htm"))

	g.NoMethod(func(c *gin.Context) {
		c.Status(http.StatusBadRequest)
	})

	g.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/")
	})
}
