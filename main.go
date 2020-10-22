package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var bucket = ""
var sess = session.Must(session.NewSession(&aws.Config{}))

func main() {
	bFlag := flag.String("bucket", "", "name of the bucket the user will use")
	uFlag := flag.String("uri", "", "set the URI in the bucket")
	pFlag := flag.String("profile", "default", "which AWS credentials should be used")
	rFlag := flag.String("region", "", "specify the AWS region")
	wFlag := flag.Int("workers", 20, "how many workers to use to download the files")
	flag.Parse()

	bucket = *bFlag
	uri := *uFlag
	profile := *pFlag
	workers := *wFlag
	region := *rFlag

	if bucket == "" || region == "" {
		flag.Usage()
		os.Exit(1)
	}

	item := uri

	sess = session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", profile),
	}))
	svc := s3.New(sess)
	res, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(item),
	})
	if err != nil {
		log.Fatalf("Error getting bucket:\n%v\n", err)
	}
	// setup concurrency -- but limited to certain number so we don't choke CPU
	finished := make(chan bool)
	files := make(chan string, len(res.Contents))

	for i := 0; i < workers; i++ {
		go worker(i, files, finished)
	}

	for _, obj := range res.Contents {
		strKey := *obj.Key
		if strKey[len(strKey)-1:] == "/" {
			if _, err := os.Stat(*obj.Key); os.IsNotExist(err) {
				os.MkdirAll(*obj.Key, os.ModePerm)
			}
		}
		files <- *obj.Key
	}
	close(files)

	// let us know when its all finished
	for fin := 1; fin <= len(res.Contents); fin++ {
		<-finished
	}
}

func downloadFile(file string) {
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Cannot open file: %v, %v", file, err)
	}

	defer f.Close()
	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(f,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(file),
		})
	if err != nil {
		log.Fatalf("Unable to download file %v, %v", file, err)
	}

	log.Println(f.Name(), "successfully downloaded with", numBytes, "bytes")
}

func join(strs ...string) string {
	var sb strings.Builder
	for _, q := range strs {
		sb.WriteString(q)
	}
	return sb.String()

}

func worker(id int, files <-chan string, finished chan<- bool) {
	for f := range files {
		if string(f[len(f)-1:]) != "/" {
			log.Println("attempting download:", f)
			downloadFile(f)
		} else {
			if _, err := os.Stat(f); os.IsNotExist(err) {
				os.MkdirAll(f, os.ModePerm)
			}
		}
		finished <- true
	}
}
