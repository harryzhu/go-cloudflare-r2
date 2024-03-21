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

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

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

	mgConn        string = GetEnv("MONGOCONN", "")
	MongoDatabase string = GetEnv("MONGODATABASE", "StableDiffusion")
	mgdb          *mongo.Database
	//mgcollection *mongo.Collection
)

const (
	IsDebug       bool   = true
	YYYYMMDD      string = "2006-01-02"
	MAX_FILE_SIZE int64  = 10 << 20
)

var (
	isDebug  bool  = true
	anyError error = errors.New("[error]")

	s3client *s3.Client
	mgclient *mongo.Client
)

func init() {
	uploadTempDir = filepath.Join(TempDir, "upload")
	log.Println("uploadTempDir:", uploadTempDir)

	MakeDirs(uploadTempDir)
	MakeDirs(LogsDir)

	//
	StartLogging()

	s3client = GetS3Client()

	_, err := s3client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		FatalError(err)
	} else {
		PrintlnDebug("found bucket")
	}

	mgclient = GetMongoClient()
	mgdb = mgclient.Database(MongoDatabase)

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
		FatalError(err)
	}

	c := s3.NewFromConfig(cfg)

	return c
}

func GetMongoClient() *mongo.Client {
	if mgConn == "" {
		mgConn = "mongodb://localhost:27017"
	}

	m, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mgConn))
	if err != nil {
		FatalError(err)
	}
	if err := m.Ping(context.TODO(), readpref.Primary()); err != nil {
		FatalError(err)
	} else {
		PrintlnDebug("mongodb connected")
	}
	return m
}

func ginOut(c *gin.Context, j JsonResponse) {
	if isDebug {
		PrintlnDebug(j.Jsonify())
	}
	c.JSON(http.StatusOK, j)
}

func StartServer() {

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	smoke := NewJsonResponse()
	smoke.WithFunction("main")
	smoke.WithStep(0)
	d := make(map[string]any, 1)
	d["status"] = "ok"
	smoke.WithData(d)

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, smoke)
	})

	api := r.Group("/api/")

	api.GET("/upload-file-form", getFormFileHTML)
	//api.POST("/upload", createFile)
	api.POST("/upload", PutFile)
	api.DELETE("/delete/:bkt/:key", deleteFile)

	r.Run(Listen)
}

func StartLogging() {
	logFile := time.Now().Format(YYYYMMDD)
	f, err := os.OpenFile(filepath.Join(LogsDir, "gin_"+logFile+".log"), os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		FatalError(err)
	}
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
}
