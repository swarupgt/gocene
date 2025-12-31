package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"gocene/config"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// All Minio/S3 APIs here

// Create new Minio Client
func CreateMinioClient() (mc *minio.Client, err error) {

	mc, err = minio.New(config.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.MinioAccessKey, config.MinioSecretKey, ""),
		Secure: false,
	})

	if err != nil {
		log.Fatalln("could not connect to minio server, err: ", err.Error())
		return
	}

	ctx := context.Background()

	exists, err := mc.BucketExists(ctx, config.MinioBucket)
	if err != nil {
		log.Fatal(err)
		return
	}

	if !exists {
		err = mc.MakeBucket(ctx, config.MinioBucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func StoreDocumentToMinio(mc *minio.Client, docID int, doc map[string]interface{}, indexName string) (err error) {

	if mc == nil {
		log.Println("no minio client passed")
		return errors.New("no minio client passed")
	}

	ctx := context.Background()
	data, err := json.Marshal(doc)
	if err != nil {
		log.Println("could not JSON marshal doc for minio store, err: ", err.Error())
		return err
	}

	objName := strings.Join([]string{config.MinioDocPathPrefix, indexName, strconv.Itoa(docID)}, "/") + ".json"

	_, err = mc.PutObject(
		ctx,
		config.MinioBucket,
		objName,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: "application/json",
		},
	)

	if err != nil {
		log.Println("could not upload doc to minio store, err: ", err.Error())
	}

	return
}

func GetDocumentFromMinio(mc *minio.Client, docID int, indexName string) (docStr string, err error) {

	ctx := context.Background()

	objName := strings.Join([]string{config.MinioDocPathPrefix, indexName, strconv.Itoa(docID)}, "/") + ".json"

	obj, err := mc.GetObject(
		ctx,
		config.MinioBucket,
		objName,
		minio.GetObjectOptions{},
	)

	defer obj.Close()

	if err != nil {
		log.Println("could not get doc from minio store, err: ", err.Error())
		return
	}

	objBytes, err := io.ReadAll(obj)
	if err != nil {
		log.Println("could not read object stream from minio get, err: ", err.Error())
		return
	}

	return string(objBytes), nil
}
