package cfr2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	//"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

var (
	accountId       string = GetEnv("CFR2ID", "")
	accessKeyId     string = GetEnv("CFR2KEYID", "")
	accessKeySecret string = GetEnv("CFR2KEYSECRET", "")

	bucketName    string = GetEnv("CFR2BUCKET", "webui")
	Listen        string = GetEnv("CFR2LISTEN", ":8080")
	TempDir       string = GetEnv("CFR2TEMPDIR", "temp")
	LogsDir       string = GetEnv("CFR2LOGSDIR", "logs")
	uploadTempDir string
)

const (
	YYYYMMDD = "2006-01-02"
)

var (
	isDebug  bool  = true
	anyError error = errors.New("[error]")
	s3client *s3.Client
)

func init() {
	logFile := time.Now().Format(YYYYMMDD)
	f, _ := os.OpenFile(filepath.Join(LogsDir, "gin_"+logFile+".log"), os.O_APPEND, 755)
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	//
	uploadTempDir = filepath.Join(TempDir, "upload")
	fmt.Println("uploadTempDir:", uploadTempDir)
	s3client = GetS3Client()

}

func GetS3Client() *s3.Client {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
	}

	c := s3.NewFromConfig(cfg)

	return c
}

func ginOut(c *gin.Context, j JsonResponse, isBreak bool) {
	if isDebug {
		log.Println(j.Jsonify())
	}
	c.JSON(http.StatusOK, j)
	if isBreak {
		c.AbortWithError(200, anyError)
	}
}

func StartServer() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	smoke := NewJsonResponse()
	d := make(map[string]any, 1)
	d["status"] = "ok"
	smoke.WithData(d)

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, smoke)
	})

	api := r.Group("/api/")

	api.GET("/upload-file-form", getFormFileHTML)
	api.POST("/upload", createFile)
	api.DELETE("/delete/:bkt/:key", deleteFile)

	r.Run(Listen)
}
