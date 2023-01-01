package s3vk

import (
	"bytes"
	xkcd "comics/basejsons"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const dirtempfiles = "temp"

var glS3Client *s3.Client

func initS3yandex() (*s3.Client, error) {
	if glS3Client != nil {
		fmt.Println("не создаёи новый client s3, используем старый")
		return glS3Client, nil
	} else {
		fmt.Println("создаём клиента s3")
	}
	// Создаем кастомный обработчик эндпоинтов, который для сервиса S3 и региона ru-msk выдаст корректный URL
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID && region == "ru-msk" {
			return aws.Endpoint{
				//PartitionID:   "yc",//не знаю что это
				URL:           "https://hb.bizmrg.com",
				SigningRegion: "ru-msk",
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})
	// Подгружаем конфигрурацию из ~/.aws/*
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(customResolver))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Создаем клиента для доступа к хранилищу S3
	client := s3.NewFromConfig(cfg)
	glS3Client = client
	return client, nil
}
func getListBuckets() error {
	client, err := initS3yandex()
	if err != nil {
		fmt.Printf("ошибка подключения к s3 %v\n", err)
		return err
	}
	// Запрашиваем список бакетов
	result, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, bucket := range result.Buckets {
		log.Printf("backet=%s creation time=%s", aws.ToString(bucket.Name), bucket.CreationDate.Format("2006-01-02 15:04:05 Monday"))
		log.Printf("backet=%s creation time=%s", aws.ToString(bucket.Name), bucket.CreationDate)
	}
	return nil
}
func getImage(addressImage string) (*bytes.Reader, int64, error) {
	if strings.HasPrefix(addressImage, "https://") == false {
		addressImage = "https://" + addressImage
	}
	req, err := http.Get(addressImage)
	if err != nil {
		return nil, 0, err
	}
	defer req.Body.Close()
	var buf bytes.Buffer
	lenlen, err := io.Copy(&buf, req.Body)
	if err != nil {
		fmt.Println(err)
		return nil, 0, err
	}
	fmt.Println("lenele", buf.Len())
	return bytes.NewReader(buf.Bytes()), lenlen, nil
}
func WriteComics(comics xkcd.TComics, nameBucket string) error {
	var metaData map[string]string
	metaDataDryData := make(map[string]string)
	metaData = make(map[string]string)
	jsonBytes, err := json.Marshal(comics)
	if err != nil {
		fmt.Printf("ошибка %v\n", err)
		return err
	}
	err = json.Unmarshal(jsonBytes, &metaDataDryData)
	if err != nil {
		strErr := fmt.Sprintf("%v", err)
		if !strings.Contains(strErr, "cannot unmarshal number into Go value of type string") {
			fmt.Printf("ошибка %v\n", err)
			return err
		}
	}
	for k, v := range metaDataDryData {
		metaData[k] = strconv.Quote(v)
	}
	jpgfile, countBytes, err := getImage(comics.Img)
	if err != nil {
		fmt.Printf("не удалось получить картинку комикса %v\n", err)
		return err
	}
	client, err := initS3yandex()
	if err != nil {
		fmt.Printf("ошибка подключения к s3 %v\n", err)
		return err
	}
	getListBuckets()
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(nameBucket),
		Key:    aws.String(strconv.Itoa(comics.Num)),
		Body:   jpgfile,
		//ContentLength: stat.Size(),
		//ContentLength: int64(len(jsonBytes)),
		//ContentLength: stat.Size(),
		ContentLength: countBytes,
		Metadata:      metaData,
	})
	if err != nil {
		fmt.Printf("ошибка %v\n", err)
		return err
	}
	return nil
}
func GetComics(nameBucket, field, val string) ([]xkcd.TComics, error) {
	var comices []xkcd.TComics
	var comics *xkcd.TComics
	client, err := initS3yandex()
	if err != nil {
		fmt.Printf("ошибка подключения к s3 %v\n", err)
		return comices, err
	}
	listObjsResponse, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(nameBucket),
		Prefix: aws.String(""),
	})
	if err != nil {
		panic("Couldn't list bucket contents")
		return comices, err
	}
	for _, object := range listObjsResponse.Contents {
		fmt.Printf("%s (%d bytes, class %v) \n", *object.Key, object.Size, object.StorageClass)
		comics = new(xkcd.TComics)
		comices = append(comices, *comics)
	}

	return comices, nil
}
func ClearBucket(nameBucket string) error {
	client, err := initS3yandex()
	if err != nil {
		fmt.Printf("ошибка подключения к s3 %v\n", err)
		return err
	}
	listObjectsV2Response, err := client.ListObjectsV2(context.TODO(),
		&s3.ListObjectsV2Input{
			Bucket: aws.String(nameBucket),
		})

	for {

		if err != nil {
			panic("Couldn't list objects...")
			return err
		}
		for _, item := range listObjectsV2Response.Contents {
			fmt.Printf("- Deleting object %s\n", *item.Key)
			_, err = client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
				Bucket: aws.String(nameBucket),
				Key:    item.Key,
			})
			if err != nil {
				panic("Couldn't delete items")
				return err
			}
		}
		if listObjectsV2Response.IsTruncated {
			listObjectsV2Response, err = client.ListObjectsV2(context.TODO(),
				&s3.ListObjectsV2Input{
					Bucket:            aws.String(nameBucket),
					ContinuationToken: listObjectsV2Response.ContinuationToken,
				})
		} else {
			break
		}

	}
	return nil
}
