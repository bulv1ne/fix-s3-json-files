package internal

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"iter"
)

func S3ClientFromProfile(profileName string) *s3.Client {
	var optFns []func(*config.LoadOptions) error
	if profileName != "" {
		optFns = append(optFns, config.WithSharedConfigProfile(profileName))
	}
	sdkConfig, err := config.LoadDefaultConfig(
		context.TODO(),
		optFns...,
	)
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		panic(err)
	}
	return s3.NewFromConfig(
		sdkConfig,
		func(o *s3.Options) {
			o.DisableLogOutputChecksumValidationSkipped = true
		},
	)
}

func ListObjectsV2(ctx context.Context, s3Client *s3.Client, params *s3.ListObjectsV2Input) iter.Seq[types.Object] {
	return func(yield func(types.Object) bool) {
		kwargs := *params
		for {
			result, err := s3Client.ListObjectsV2(ctx, &kwargs)
			if err != nil {
				fmt.Println("Error listing objects:", err)
				return
			}
			if !yieldSlice(yield, result.Contents) {
				return
			}

			if result.NextContinuationToken != nil {
				kwargs.ContinuationToken = result.NextContinuationToken
			} else {
				break
			}
		}
	}
}

func yieldSlice[T any](yield func(T) bool, slice []T) bool {
	for _, item := range slice {
		if !yield(item) {
			return false
		}
	}
	return true
}
