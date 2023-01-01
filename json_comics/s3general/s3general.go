package s3general

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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/ о том как использовать несколько профилей
// о формате файлов credentials и config в папке ~/.aws https://docs.aws.amazon.com/sdkref/latest/guide/file-format.html
// документация по go SDK для aws https://docs.aws.amazon.com/sdk-for-go/api/service/s3/
// документация ещё от aws общая https://docs.aws.amazon.com/sdk-for-go/api/
// документация  по пакету service/s3 https://docs.aws.amazon.com/sdk-for-go/api/service/s3/
// документация по пакету aws https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws
// документация по пакету config https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/config
const dirtempfiles = "temp"

var glS3Client *s3.Client

func initS3general() (*s3.Client, error) {
	if glS3Client != nil {
		fmt.Println("не создаёи новый client s3, используем старый")
		return glS3Client, nil
	} else {
		fmt.Println("создаём клиента s3")
	}
	// Создаем кастомный обработчик эндпоинтов, который для сервиса S3 и региона ru-central1 выдаст корректный URL

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		fmt.Printf("service == %v; region == %v\n", service, region)
		if service == s3.ServiceID && region == "ru-central1" {
			return aws.Endpoint{
				PartitionID: "yc",
				//URL:           "https://hb.bizmrg.com",
				//SigningRegion: "ru-msk",
				URL:           "https://storage.yandexcloud.net",
				SigningRegion: "ru-central1",
			}, nil
		}
		if service == s3.ServiceID && region == "ru-msk" {
			return aws.Endpoint{
				//PartitionID: "yc",
				URL:           "https://hb.bizmrg.com",
				SigningRegion: "ru-msk",
				//URL:           "https://storage.yandexcloud.net",
				//SigningRegion: "ru-central1",
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested, %s", options)
	})
	// Подгружаем конфигрурацию из ~/.aws/*
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(customResolver), config.WithSharedConfigProfile("vk"))
	//cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(customResolver), config.WithSharedConfigProfile("yandex"))
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
	client, err := initS3general()
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
func FindComics(text string) ([]xkcd.TComics, []bytes.Buffer /*[]*bytes.Reader*/, error) {
	//1 способ
	//получем впсиок всех объектов в бакете
	//для каждого объекта из бакета получаем метаданные
	//для каждого метаданного проверяем не содержит ли он строку поиска
	//и если содержит добавялем номер найденного
	//для каждого соответсвующего номер выводим поле из метаданных занголовок и transcriptions
	//а так же выводим картинку этого комикса
	//2 способ
	//поле transcription хранится в отдельной базе типа elasticseaarch по этой базе и делаем поиск
	//из elasticsearch получеме все ключи
	//по этим ключам получаем из базы s3 картинки и заголовки

	//общее
	//для каждого найденного коммикса получаем jpg и сохраняем во временную папку или массив редеров
	return nil, nil, nil
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
	client, err := initS3general()
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
	client, err := initS3general()
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
	client, err := initS3general()
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
