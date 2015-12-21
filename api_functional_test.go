package minio_test

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/minio/minio-go"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz01234569"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randString(n int, src rand.Source) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b[0:30])
}

func TestFunctional(t *testing.T) {
	c, err := minio.New("play.minio.io:9002",
		"Q3AM3UQ867SPQQA43P2F", "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG", false)
	if err != nil {
		t.Fatal("Error:", err)
	}

	// Set user agent.
	c.SetAppInfo("Test", "0.1.0")

	bucketName := randString(60, rand.NewSource(time.Now().UnixNano()))
	err = c.MakeBucket(bucketName, "private", "us-east-1")
	if err != nil {
		t.Fatal("Error:", err, bucketName)
	}

	err = c.BucketExists(bucketName)
	if err != nil {
		t.Fatal("Error:", err, bucketName)
	}

	err = c.SetBucketACL(bucketName, "public-read-write")
	if err != nil {
		t.Fatal("Error:", err)
	}

	acl, err := c.GetBucketACL(bucketName)
	if err != nil {
		t.Fatal("Error:", err)
	}
	if acl != minio.BucketACL("public-read-write") {
		t.Fatal("Error:", acl)
	}

	_, err = c.ListBuckets()
	if err != nil {
		t.Fatal("Error:", err)
	}

	objectName := bucketName + "Minio"
	readSeeker := bytes.NewReader([]byte("Hello World!"))

	n, err := c.PutObject(bucketName, objectName, readSeeker, int64(readSeeker.Len()), "")
	if err != nil {
		t.Fatal("Error: ", err)
	}
	if n != int64(len([]byte("Hello World!"))) {
		t.Fatal("Error: bad length ", n, readSeeker.Len())
	}

	newReadSeeker, _, err := c.GetObject(bucketName, objectName)
	if err != nil {
		t.Fatal("Error: ", err)
	}

	newReadBytes, err := ioutil.ReadAll(newReadSeeker)
	if err != nil {
		t.Fatal("Error: ", err)
	}

	if !bytes.Equal(newReadBytes, []byte("Hello World!")) {
		t.Fatal("Error: bytes invalid.")
	}

	err = c.RemoveObject(bucketName, objectName)
	if err != nil {
		t.Fatal("Error: ", err)
	}
	err = c.RemoveBucket(bucketName)
	if err != nil {
		t.Fatal("Error:", err)
	}

	err = c.RemoveBucket("bucket1")
	if err == nil {
		t.Fatal("Error:")
	}

	if err.Error() != "The specified bucket does not exist." {
		t.Fatal("Error:", err)
	}

}
