package cfr2

import (
	"bytes"
	"path/filepath"

	//"errors"
	"io"

	//"io/ioutil"

	"os"

	"context"
	//"os"

	"github.com/gin-gonic/gin"

	//"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3client *s3.Client

func uploadFile(c *gin.Context) {
	f, _ := c.FormFile("file")

	dst := filepath.Join(uploadTempDir, f.Filename)
	fSize := f.Size
	c.SaveUploadedFile(f, dst)

	fMD5, err := MD5File(dst)
	if err != nil {
		log.Println("cannot get md5")
		c.Abort()
	}

	fKey := fMD5

	if Exists(bucketName, fKey) {
		log.Println("SKIP(as exists) ...")
		c.Abort()
	} else {
		// upload
		fi, err := os.Open(dst)
		if err != nil {
			log.Println(err)
			c.Abort()
		}

		s3uploader := manager.NewUploader(s3client)
		fBytes, _ := io.ReadAll(fi)

		result, err := s3uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket:        aws.String(bucketName),
			Key:           aws.String(fKey),
			ContentType:   aws.String("image/png"),
			ContentLength: aws.Int64(fSize),
			Body:          bytes.NewReader(fBytes),
		})

		if err != nil {
			log.Println(err)
		} else {
			os.Remove(dst)
			log.Println("UPLOADED:", result)
		}
	}

}

func deleteFile(c *gin.Context) {
	bkt := c.Param("bkt")
	key := c.Param("key")

	if bkt != "" && key != "" {
		log.Println(bkt, key)
		if Exists(bkt, key) {
			log.Println("DELETE ...")
			result, err := s3client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: aws.String(bkt),
				Key:    aws.String(key),
			})
			if err != nil {
				log.Println(err)
			} else {
				log.Println(result)
			}
		} else {
			log.Println("key does not exist:", bkt, "/", key)
		}
	}

}

func Exists(b string, k string) bool {
	_, err := s3client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(b),
		Key:    aws.String(k),
	})

	if err != nil {
		return false
	}
	return true
}

func init() {

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

	s3client = s3.NewFromConfig(cfg)
}
