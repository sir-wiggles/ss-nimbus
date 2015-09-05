package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"strings"
	"sync"
	"time"
)

func NewExpedition(num_of_scouts int) *Expedition {

	return &Expedition{
		Result:  make(chan string, 1),
		Disband: make(chan bool, 1),
	}
}

type Expedition struct {
	sync.RWMutex
	Result  chan string
	Disband chan bool
	Nothing int64
}

func (e *Expedition) Explore(group string, request, response chan string, wg *sync.WaitGroup) {

	defer wg.Done()
	// send the connection back to the pool when done with it

	var (
		buckets    []string
		sWatchChan = make(chan string, 1)
		wWatchChan = make(chan bool, 1)
		swg        sync.WaitGroup
	)

	switch group {
	case "blocklist":
		buckets = blocklistBuckets
	case "block":
		buckets = blockBuckets
	}

	for id := range request {
		scoutReport := make(chan string, 1)
		for _, bucket := range buckets {
			swg.Add(1)
			go e.Scout(bucket, id, scoutReport, &swg)
		}

		// case 1 of 2: All workers finished without finding the block anywhere
		go func(c chan bool) {
			swg.Wait()
			time.Sleep(1000 * time.Millisecond)
			c <- true
		}(wWatchChan)

		// case 2 of 2: A worker found a result, send that back now.
		go func(c chan string) {
			result := <-scoutReport
			c <- result
		}(sWatchChan)

		select {
		// case 1 of 2 is a 404
		case <-wWatchChan:
			// case 2 of 2 we found something return now
		case location := <-sWatchChan:
			response <- location
		}
	}
	swg.Wait()
}

func (e *Expedition) Scout(directions, id string, r chan string, wg *sync.WaitGroup) {

	defer wg.Done()

	steps := strings.Split(directions, ";")
	region := steps[1]
	bucket := steps[0]

	conn := pool.Get(region)
	defer func(c *s3.S3, r string) {
		pool.Put(c, r)
	}(conn, region)

	_, err := conn.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(id),
	})
	if err != nil {
		return
	}
	r <- fmt.Sprintf("%s:  %s", id, bucket)
}
