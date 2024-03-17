package cfr2

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	bucketName      string = "webui"
	accountId       string = GetEnv("CFR2ID", "")
	accessKeyId     string = GetEnv("CFR2KEYID", "")
	accessKeySecret string = GetEnv("CFR2KEYSECRET", "")

	uploadTempDir string = "./temp/upload"
)

func StartServer() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	api := r.Group("/api/")

	api.POST("/upload", uploadFile)
	api.DELETE("/delete/:bkt/:key", deleteFile)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
