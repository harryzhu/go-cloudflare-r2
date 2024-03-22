package cfr2

import (
	// "fmt"
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
	upload_file_category  string = "upload-file-category"
)

func PutFile(c *gin.Context) {
	resp := NewJsonResponse()
	resp.WithFunction("createFile")

	fUpload, err := c.FormFile(upload_file_form_name)
	if err != nil {
		PrintlnError(err)
		return
	}

	user_name := c.PostForm(upload_user_name)
	file_category := c.PostForm(upload_file_category)
	path_prefix := c.PostForm(upload_path_prefix)

	fDst := filepath.Join(uploadTempDir, fUpload.Filename)
	err = c.SaveUploadedFile(fUpload, fDst)
	if err != nil {
		PrintlnError(err)
		return
	}

	ett := NewEntity(fDst)

	fKey := strings.Join([]string{path_prefix, file_category, user_name, ett.MD5 + ett.Ext}, "/")
	fKey = strings.ToLower(fKey)
	PrintlnDebug(fKey)

	ett.WithString("User", user_name)
	ett.WithString("Key", fKey)
	ett.WithString("Category", file_category)

	ett.SaveS3()

	ett.SaveKVDB("images")

	ginOut(c, resp)
}

func GetFile(c *gin.Context) {
	k := ""
	item := NewItem(k)
	item.GetFrom("images")
}

func ListFileByUser(c *gin.Context) {
	username := strings.TrimPrefix(c.Param("username"), "/")

	item := &Item{}
	item.GetFrom("images")
}

func DeleteFile(c *gin.Context) {
	resp := NewJsonResponse()
	resp.WithFunction("deleteFile")

	bkt := strings.TrimPrefix(c.Param("bkt"), "/")
	key := strings.TrimPrefix(c.Param("key"), "/")
	PrintlnDebug("bkt:", bkt)
	PrintlnDebug("key:", key)
	if bkt != "" && key != "" {
		ett := &Entity{}
		ett.Error = nil
		ett.Bucket = bkt
		ett.User = "u0"
		ett.Category = "model_22"
		ett.Key = key

		ett.DeleteFromS3()
	}
	ginOut(c, resp)
}

func getFormFileHTML(c *gin.Context) {
	h := gin.H{
		"upload_file_form_name": upload_file_form_name,
		"upload_path_prefix":    upload_path_prefix,
		"upload_user_name":      upload_user_name,
		"upload_file_category":  upload_file_category,
	}
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
