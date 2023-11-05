package s3

import (
	"context"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/tukaelu/ikesu/internal/config/loader"
)

const defaultRegion = "ap-northeast-1"

func init() {
	loader.Register("s3", &Loader{})
}

type Loader struct{}

func (d *Loader) LoadWithContext(ctx context.Context, u *url.URL) ([]byte, error) {
	client, err := newS3Client(ctx, u.Host, resolveRegion(u))
	if err != nil {
		return nil, err
	}
	return fetchFromBucket(ctx, client, u.Host, u.Path)
}

func resolveRegion(u *url.URL) string {
	// e.g. s3://bucket/key?regionHint=ap-northeast-1
	if rh := strings.TrimSpace(u.Query().Get("regionHint")); rh != "" {
		return rh
	}
	return defaultRegion
}

func newS3Client(ctx context.Context, bucket, regionHint string) (*s3manager.Downloader, error) {
	sess := session.Must(session.NewSession())

	r, err := s3manager.GetBucketRegion(ctx, sess, bucket, regionHint)
	if err != nil {
		return nil, err
	}
	sess.Config.Region = aws.String(r)

	return s3manager.NewDownloader(sess), nil
}

func fetchFromBucket(ctx context.Context, client *s3manager.Downloader, bucket, key string) ([]byte, error) {
	buf := &aws.WriteAtBuffer{}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if _, err := client.Download(buf, input); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
