package cfr2

import (
	"bytes"
	"fmt"
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
	PrintlnDebug(f.Filename)

	dst := filepath.Join(uploadTempDir, f.Filename)
	fSize := f.Size
	err := c.SaveUploadedFile(f, dst)
	if err != nil {
		PrintlnError(err)
		resp.AutoStep()
		resp.WithErrorMessage(500, err.Error())
		ginOut(c, resp)
		return
	}

	fMD5, err := MD5File(dst)
	if err != nil {
		resp.AutoStep()
		resp.WithErrorMessage(500, err.Error())
		ginOut(c, resp)
		return
	}

	fKey := fMD5

	if Exists(bucketName, fKey) {
		resp.AutoStep()
		resp.WithErrorMessage(0, "SKIP save as exist")
		ginOut(c, resp)

		PrintlnDebug("SKIP save as exist")
		return
	} else {
		// upload
		fi, err := os.Open(dst)
		if err != nil {
			resp.WithErrorMessage(500, err.Error())
			resp.AutoStep()
			ginOut(c, resp)
			return
		}

		s3uploader := manager.NewUploader(s3client)
		fBytes, _ := io.ReadAll(fi)

		_, err = s3uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket:        aws.String(bucketName),
			Key:           aws.String(fKey),
			ContentType:   aws.String("image/png"),
			ContentLength: aws.Int64(fSize),
			Body:          bytes.NewReader(fBytes),
		})

		if err != nil {
			PrintlnError(err)
		} else {
			err = os.Remove(dst)
			if err != nil {
				resp.AutoStep()
				resp.WithErrorMessage(500, err.Error())
			}
			PrintlnDebug("R2 UPLOADED:")
		}
	}

	ginOut(c, resp)
}

func deleteFile(c *gin.Context) {
	resp := NewJsonResponse()
	resp.WithFunction("deleteFile")

	bkt := c.Param("bkt")
	key := c.Param("key")

	if bkt != "" && key != "" {
		PrintlnDebug(fmt.Sprintf("%s/%s", bkt, key))
		if Exists(bkt, key) {
			PrintlnDebug("DELETE ...")
			_, err := s3client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: aws.String(bkt),
				Key:    aws.String(key),
			})
			if err != nil {
				resp.AutoStep()
				resp.WithErrorMessage(500, err.Error())
				PrintlnError(err)
			} else {
				resp.AutoStep()
				resp.WithErrorMessage(0, "ok")
				PrintlnDebug(fmt.Sprintf("DELETED: %s/%s", bkt, key))
			}
		} else {
			resp.AutoStep()
			resp.WithErrorMessage(404, fmt.Sprintf("key does not exist: %s/%s", bkt, key))
			PrintlnDebug(fmt.Sprintf("Error: key does not exist: %s/%s", bkt, key))
		}
	} else {
		resp.AutoStep()
		resp.WithErrorMessage(500, "bucket or key cannot be empty")
	}
	ginOut(c, resp)
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
		PrintlnError(err)
		return false
	}
	return true
}
