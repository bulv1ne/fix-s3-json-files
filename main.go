package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fix_s3_json_files/internal"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

const bucketName = "TODO"  // Replace with your bucket name
const prefixName = "TODO"  // Replace with your prefix name e.g. "raw/domain/live/"
const profileName = "TODO" // Replace with your AWS profile name

func main() {
	fs := ff.NewFlagSet("aws_s3_sync")
	prefix := fs.StringLong("prefix", "", "prefix")
	realRun := fs.BoolLong("realrun", "Real run the operation")

	err := ff.Parse(
		fs,
		os.Args[1:],
		ff.WithEnvVarPrefix("AWS_S3_SYNC"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
	)
	if err != nil {
		fmt.Printf("%s\n", ffhelp.Flags(fs))
		fmt.Printf("err=%v\n", err)
		return
	}
	*prefix = prefixName + *prefix
	*prefix, _ = strings.CutSuffix(*prefix, "/")
	bucket := bucketName

	client := internal.S3ClientFromProfile(profileName)

	processFiles(client, bucket, *prefix, *realRun)
}

func processFiles(client *s3.Client, bucket string, prefix string, realRun bool) {
	fmt.Printf("Processing files in s3://%s/%s\n", bucket, prefix)

	ctx := context.Background()
	if !realRun {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
	}

	var wg sync.WaitGroup
	jobs := make(chan func())
	maxWorkers := 1
	if realRun {
		maxWorkers = 10
	}
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go Worker(&wg, jobs)
	}
	QueueJobs(ctx, client, bucket, prefix, realRun, jobs)
	close(jobs)
	wg.Wait()
}

// QueueJobs lists S3 objects with the given prefix and enqueues jobs for processing them.
func QueueJobs(ctx context.Context, client *s3.Client, bucket string, prefix string, realRun bool, jobs chan<- func()) {
	for object := range internal.ListObjectsV2(context.Background(), client, &s3.ListObjectsV2Input{Bucket: &bucket, Prefix: &prefix}) {
		select {
		case <-ctx.Done():
			return
		default:
			// Do nothing
		}
		if strings.HasSuffix(*object.Key, "/") {
			continue
		}
		jobs <- func() {
			_ = ProcessS3Object(context.Background(), client, bucket, object, realRun)
		}
	}
}

// Worker runs jobs from the jobs channel until it is closed.
// It signals completion via the WaitGroup.
func Worker(wg *sync.WaitGroup, jobs <-chan func()) {
	defer wg.Done()
	for job := range jobs {
		job := job
		job()
	}
}

// ProcessS3Object downloads, decompresses, and attempts to fix a gzipped JSON object from S3.
// If changes are detected, it uploads the fixed object back to S3 (if realRun is true),
// or prints the changes otherwise.
func ProcessS3Object(ctx context.Context, client *s3.Client, bucket string, object types.Object, realRun bool) error {
	fmt.Println("Checking object", *object.Key)
	result, err := client.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucket, Key: object.Key})
	// Reader section
	if err != nil {
		fmt.Println("Coult not get object", *object.Key)
		return err
	}
	defer internal.CloseAndLogOnError(result.Body)

	reader, err := gzip.NewReader(result.Body)
	if err != nil {
		fmt.Println("Could not read object")
		return err
	}
	defer internal.CloseAndLogOnError(reader)

	// Writer section
	output := &bytes.Buffer{}
	writer := gzip.NewWriter(output)
	bufferedWriter := bufio.NewWriter(writer)

	changes, err := internal.FixJsonStream(reader, bufferedWriter)
	if err != nil {
		fmt.Println("Could not fix json", err)
		return err
	}
	if err = writer.Flush(); err != nil {
		fmt.Println("Could not flush the buffer", err)
		return err
	}
	// Close the writer
	if err = writer.Close(); err != nil {
		fmt.Println("Could not close gzip writer", err)
		return err
	}

	if changes {
		fmt.Println("Changes detected in", *object.Key)
		if realRun {
			_, err = client.PutObject(
				ctx,
				&s3.PutObjectInput{Bucket: &bucket, Key: object.Key, Body: bytes.NewReader(output.Bytes())},
			)
			if err != nil {
				fmt.Println("Could not put object", *object.Key, err)
			}
		} else {
			// Verify the changes
			r, _ := gzip.NewReader(bytes.NewReader(output.Bytes()))
			data, _ := io.ReadAll(r)
			fmt.Printf("Changes for key: %s\n%s\n", *object.Key, string(data))
		}
	}
	return nil
}
