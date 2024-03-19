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
	//"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	//"github.com/aws/aws-sdk-go-v2/config"
	//"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	upload_file_form_name string = "upload-file-form"
)

func createFile(c *gin.Context) {
	resp := NewJsonResponse()
	resp.WithFunction("createFile")

	f, _ := c.FormFile(upload_file_form_name)
	log.Println(f.Filename)
	dst := filepath.Join(uploadTempDir, f.Filename)
	fSize := f.Size
	c.SaveUploadedFile(f, dst)

	fMD5, err := MD5File(dst)
	if err != nil {
		resp.AutoStep()
		resp.WithErrorMessage(500, err.Error())
		ginOut(c, resp, true)
	}

	fKey := fMD5

	if Exists(bucketName, fKey) {
		resp.AutoStep()
		ginOut(c, resp, true)
		log.Println("SKIP(as exists) ...")
	} else {
		// upload
		fi, err := os.Open(dst)
		if err != nil {
			resp.WithErrorMessage(500, err.Error())
			resp.AutoStep()
			ginOut(c, resp, true)
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
			err = os.Remove(dst)
			if err != nil {
				resp.AutoStep()
				resp.WithErrorMessage(500, err.Error())
			}
			log.Println("R2 UPLOADED:", result)
		}
	}

	ginOut(c, resp, true)
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

func getFormFileHTML(c *gin.Context) {
	h := gin.H{"upload_file_form_name": upload_file_form_name}
	c.HTML(200, "upload-file-form.html", h)
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
