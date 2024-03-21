package cfr2

import (
	"fmt"
	"path/filepath"
	"strings"

	//"errors"

	//"io/ioutil"

	"context"
	//"os"

	"github.com/gin-gonic/gin"

	//"encoding/json"
	//"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	//"github.com/aws/aws-sdk-go-v2/config"
	//"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	upload_file_form_name string = "upload-file-form"
	upload_path_prefix    string = "upload-path-prefix"
	upload_user_name      string = "upload-user-name"
)

func PutFile(c *gin.Context) {
	resp := NewJsonResponse()
	resp.WithFunction("createFile")

	fUpload, err := c.FormFile(upload_file_form_name)
	if err != nil {
		PrintlnError(err)
		return
	}

	path_prefix := c.PostForm(upload_path_prefix)
	user_name := c.PostForm(upload_user_name)

	fDst := filepath.Join(uploadTempDir, fUpload.Filename)
	err = c.SaveUploadedFile(fUpload, fDst)
	if err != nil {
		PrintlnError(err)
		return
	}

	ett := NewEntity(fDst)

	fKey := strings.Join([]string{path_prefix, user_name, ett.MD5 + ett.Ext}, "/")
	PrintlnDebug(fKey)

	ett.WithString("User", user_name)
	ett.WithString("Key", fKey)

	ett.SaveS3()

	ett.SaveKVDB()

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

func HeadFile(b string, k string) (*s3.HeadObjectOutput, error) {
	h, err := s3client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(b),
		Key:    aws.String(k),
	})

	if err != nil {
		PrintlnError(err)
		return &s3.HeadObjectOutput{}, err
	}
	return h, nil
}

// ===============
