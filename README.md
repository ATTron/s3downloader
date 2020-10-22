# S3 Downloader
Download entire s3 buckets quickly with the power of goroutines!

## Usage
```
./s3download 
Usage of ./s3download:
  -bucket string
        name of the bucket the user will use
  -profile string
        which AWS credentials should be used (default "default")
  -region string
        specify the AWS region
  -uri string
        set the URI in the bucket
  -workers int
        how many workers to use to download the files (default 20)
```

Example downloading entire bucket named 'my-bucket' with aws profile attron (s3://my-bucket):  
```bash
./s3download -bucket my-bucket -profile attron -region us-west-1
```

Example downloading direcotry '/testing/example' from bucket named 'my-bucket' with aws profile attron and 15 workers (s3://my-bucket/tesing/example):
```bash
./s3download -bucket my-bucket -uri testing/example -profile attron -region us-west-1 -workers 15
```

Your bucket / URI will be downloaded at the directory where this script is run