package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create an AWS resource (S3 Bucket)
		bucketResourceName := "contentful-app-bucket"
		bucket, err := s3.NewBucket(ctx, bucketResourceName, &s3.BucketArgs{
			Bucket: pulumi.String(bucketResourceName),
		})
		if err != nil {
			return err
		}

		// Bucket public access block
		bucketBlockPublicAccessResourceName := "bucket-block-public-access"
		_, err = s3.NewBucketPublicAccessBlock(ctx, bucketBlockPublicAccessResourceName, &s3.BucketPublicAccessBlockArgs{
			Bucket:                bucket.ID(),
			BlockPublicAcls:       pulumi.Bool(true),
			BlockPublicPolicy:     pulumi.Bool(true),
			IgnorePublicAcls:      pulumi.Bool(true),
			RestrictPublicBuckets: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		// Create OriginAccessIdentity
		oaiName := "origin-access-identity-s3-bucket"
		originAccessIdentity, err := cloudfront.NewOriginAccessIdentity(ctx, oaiName, &cloudfront.OriginAccessIdentityArgs{
			Comment: pulumi.String("Origin access identity for s3 bucket access"),
		})
		if err != nil {
			return err
		}

		// Create CloudFront distribution
		cloudfrontDistName := "contentful-app-cdn"
		bucketOriginId := "contentful-app-bucket-origin"
		cloudFrontDistribution, err := cloudfront.NewDistribution(ctx, cloudfrontDistName, &cloudfront.DistributionArgs{
			Enabled: pulumi.Bool(true),
			Origins: cloudfront.DistributionOriginArray{
				&cloudfront.DistributionOriginArgs{
					OriginId:   pulumi.String(bucketOriginId),
					DomainName: bucket.BucketDomainName,
					S3OriginConfig: &cloudfront.DistributionOriginS3OriginConfigArgs{
						OriginAccessIdentity: originAccessIdentity.CloudfrontAccessIdentityPath,
					},
				},
			},
			DefaultCacheBehavior: &cloudfront.DistributionDefaultCacheBehaviorArgs{
				AllowedMethods: pulumi.StringArray{
					pulumi.String("GET"),
					pulumi.String("HEAD"),
				},
				CachedMethods: pulumi.StringArray{
					pulumi.String("GET"),
					pulumi.String("HEAD"),
				},
				TargetOriginId: pulumi.String(bucketOriginId),
				ForwardedValues: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesArgs{
					QueryString: pulumi.Bool(false),
					Cookies: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesCookiesArgs{
						Forward: pulumi.String("none"),
					},
				},
				ViewerProtocolPolicy: pulumi.String("allow-all"),
			},
			Restrictions: &cloudfront.DistributionRestrictionsArgs{
				GeoRestriction: cloudfront.DistributionRestrictionsGeoRestrictionArgs{
					RestrictionType: pulumi.String("none"),
				},
			},
			ViewerCertificate: cloudfront.DistributionViewerCertificateArgs{
				CloudfrontDefaultCertificate: pulumi.Bool(true),
			},
			// TODO set DefaultRootObject
			DefaultRootObject: pulumi.String("index.html"),
		})
		if err != nil {
			return err
		}

		// Create bucket policy which provides access to bucket objects via CloudFront
		bucketPolicyName := "PolicyForCloudFrontPrivateContent"
		bucketPolicy, err := s3.NewBucketPolicy(ctx, bucketPolicyName, &s3.BucketPolicyArgs{
			Bucket: bucket.ID(),
			Policy: pulumi.All(bucket.Arn, originAccessIdentity.IamArn).ApplyT(func(args []interface{}) (string, error) {
				bucketPolicyJson, err := PublicReadPolicyForBucket(bucketPolicyName, args[0].(string), args[1].(string))
				if err != nil {
					emptyString := ""
					return emptyString, err
				}
				return string(bucketPolicyJson), nil
			}).(pulumi.StringOutput),
		})
		if err != nil {
			return err
		}

		// Export the name of the bucket
		ctx.Export("bucketName", bucket.ID())

		// Export CloudFront domain name
		ctx.Export("cloudFrontDistributionDomainName", cloudFrontDistribution.DomainName)

		// Export bucket policy
		ctx.Export("bucketPolicy", bucketPolicy.Policy)

		return nil
	})
}

func PublicReadPolicyForBucket(bucketPolicyName string, bucketArn string, originAccessIdentityIamArn string) ([]byte, error) {
	jsonToReturn, err := json.Marshal(map[string]interface{}{
		"Version": "2012-10-17",
		"Id":      bucketPolicyName,
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					"AWS": originAccessIdentityIamArn,
				},
				"Action":   "s3:GetObject",
				"Resource": fmt.Sprintf("%s/*", bucketArn),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return jsonToReturn, nil
}
