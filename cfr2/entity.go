package cfr2

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Entity struct {
	User        string
	Bucket      string
	Category    string
	LocalPath   string
	Key         string
	MD5         string
	Ext         string
	Perm        string
	Size        int64
	ContentType string
	IsPublic    int
	Error       error
	Data        io.Reader
}

func NewEntity(fpath string) *Entity {
	finfo, err := os.Stat(fpath)
	if err != nil {
		PrintlnError(err)
		return nil
	}

	if finfo.Size() > MAX_FILE_SIZE {
		PrintlnDebug("Error: the file size exceeds configured limit", MAX_FILE_SIZE)
		return nil
	}

	absfpath, err := filepath.Abs(fpath)
	if err != nil {
		PrintlnError(err)
		return nil
	}

	fext := filepath.Ext(absfpath)
	PrintlnDebug(fext)
	if fext == "" {
		PrintlnDebug("cannot upload no-extension file")
		return nil
	}

	contentType := mime.TypeByExtension(fext)
	if contentType == "" {
		PrintlnDebug("cannot parse mime type of the file")
		return nil
	}

	md5absfpath, err := MD5File(absfpath)
	if err != nil {
		PrintlnError(err)
		return nil
	}

	fh, err := os.Open(absfpath)
	if err != nil {
		PrintlnError(err)
		return nil
	}

	fBytes, _ := io.ReadAll(fh)

	ett := &Entity{}
	ett.Bucket = bucketName
	ett.LocalPath = absfpath
	ett.Size = finfo.Size()
	ett.MD5 = md5absfpath
	ett.Ext = fext
	ett.ContentType = contentType
	ett.Data = bytes.NewReader(fBytes)
	return ett
}

func (ett *Entity) WithString(k string, v string) *Entity {
	switch strings.ToLower(k) {
	case "user":
		ett.User = v
	case "bucket":
		ett.Bucket = v
	case "category":
		ett.Category = v
	case "localpath":
		ett.LocalPath = v
	case "key":
		ett.Key = v
	case "ext":
		ett.Key = v
	case "md5":
		ett.MD5 = v
	case "contenttype":
		ett.ContentType = v
	case "perm":
		ett.Perm = v
	}
	return ett
}

func (ett *Entity) WithInt64(k string, v int64) *Entity {
	switch strings.ToLower(k) {
	case "size":
		ett.Size = v
	}

	return ett
}

func (ett *Entity) WithInt(k string, v int) *Entity {
	switch strings.ToLower(k) {
	case "ispublic":
		ett.IsPublic = v
	}
	return ett
}

func (ett *Entity) SaveS3() *Entity {
	if ett.Error != nil || ett.Data == nil {
		return ett
	}

	if ett.User == "" || ett.Bucket == "" || ett.Category == "" || ett.Key == "" || ett.ContentType == "" || ett.Size == 0 {
		return ett
	}

	s3uploader := manager.NewUploader(s3client)

	_, err := s3uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(ett.Bucket),
		Key:           aws.String(ett.Key),
		ContentType:   aws.String(ett.ContentType),
		ContentLength: aws.Int64(ett.Size),
		CacheControl:  aws.String("max-age=86400"),
		Body:          ett.Data,
	})

	if err != nil {
		ett.Error = err
		PrintlnError(err)
	}

	return ett
}

func (ett *Entity) DeleteFromS3() *Entity {
	if ett.Error != nil {
		return ett
	}

	if ett.Bucket == "" || ett.Key == "" {
		ett.Error = errors.New("bucket/key cannot be empty")
		return ett
	}

	result, err := s3client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(ett.Bucket),
		Key:    aws.String(ett.Key),
	})

	if err != nil {
		ett.Error = err
		PrintlnError(err)
	} else {
		PrintlnDebug(result)
	}

	return ett
}

func (ett *Entity) SaveKVDB(coll string) *Entity {
	if ett.Error != nil || ett.Key == "" || coll == "" {
		return ett
	}
	item := NewItem(ett.Key)
	item.Size = ett.Size
	item.Owner = ett.User

	meta := make(map[string]string, 10)

	item.Meta = meta
	item.SaveTo(coll)
	return ett
}

func (ett *Entity) DeleteFromKVDB(coll string) *Entity {
	if ett.Error != nil || ett.Key == "" || coll == "" {
		return ett
	}
	item := NewItem(ett.Key)

	item.DeleteFrom(coll)
	return ett
}

func (ett *Entity) SaveRDB(k string, v int) *Entity {
	if ett.Error != nil {
		return ett
	}

	return ett
}
