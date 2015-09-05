package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Pool map[string]chan *s3.S3


func (p *Pool) Get(region string) *s3.S3 {
	return <-(*p)[region]
}

func (p *Pool) Put(conn *s3.S3, region string) {
	c := (*p)[region]
	c <- conn
}

func FillPool(pool_size int) {
	// initialize the connectino pool
	pool = make(Pool)

	for _, region := range regions {
		c := make(chan *s3.S3, pool_size)
		for i := 0; i < pool_size; i++ {
			c <- NewS3Connection(aws_key, aws_sec, region)
		}
		pool[region] = c
	}
}

// helper method to get a connection to s3
func NewS3Connection(key, secret, region string) *s3.S3 {
	config := &aws.Config{
		Region:      aws.String(region),
		MaxRetries:  aws.Int(1),
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
	}
	service := s3.New(config)
	return service
}
