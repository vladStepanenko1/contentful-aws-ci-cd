package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create an AWS resource (S3 Bucket)
		bucketResourceName := "contentful-app-bucket"
		bucket, err := s3.NewBucket(ctx, bucketResourceName, &s3.BucketArgs{
			Acl:    pulumi.String("private"),
			Bucket: pulumi.String(bucketResourceName),
		})
		if err != nil {
			return err
		}

		// Bucket public access block
		_, err = s3.NewBucketPublicAccessBlock(ctx, "bucket-block-public-access", &s3.BucketPublicAccessBlockArgs{
			Bucket:                pulumi.String(bucketResourceName),
			BlockPublicAcls:       pulumi.Bool(true),
			BlockPublicPolicy:     pulumi.Bool(true),
			IgnorePublicAcls:      pulumi.Bool(true),
			RestrictPublicBuckets: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		// Export the name of the bucket
		ctx.Export("bucketName", bucket.ID())
		return nil
	})
}
