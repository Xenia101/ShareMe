package www

import (
	"net/http"
	"path/filepath"

	"shareme/share"

	"github.com/gin-gonic/gin"
)

func initGin(g *gin.Engine) {
	g.Use(gin.Logger())
	g.Use(handlePanic)
	g.MaxMultipartMemory = 128 * 1024 // 128 KiB 이상은 모조리 임시파일로

	g.LoadHTMLGlob(filepath.Join(share.Config.Dir.Public, "*.htm"))

	g.Static("/static", filepath.Join(share.Config.Dir.Public, "static"))
	g.POST("/upload", handleUpload)
	g.POST("/download", handleDownload)

	g.StaticFile("/", filepath.Join(share.Config.Dir.Public, "index.htm"))

	g.NoMethod(func(c *gin.Context) {
		c.Status(http.StatusBadRequest)
	})

	g.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/")
	})
}
