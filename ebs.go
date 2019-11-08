package main

import (
	"os"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"k8s.io/klog"
)

func getEBSTags(annotation string) []*ec2.Tag {

	var tags []*ec2.Tag

	tagsList := strings.Split(strings.TrimSpace(annotation), ",")

	for _, v := range tagsList {

		tag := strings.Split(strings.TrimSpace(v), "=")

		if len(tag) == 2 {
			tags = append(tags, &ec2.Tag{
				Key:   aws.String(tag[0]),
				Value: aws.String(tag[1]),
			})
		} else if len(tag) == 1 {
			tags = append(tags, &ec2.Tag{
				Key:   aws.String(tag[0]),
				Value: aws.String(""),
			})
		} else {
			klog.Warningf("Invalid annotation %q:", annotation)
			return nil
		}
	}

	return tags

}

func createTags(volume *string, tags []*ec2.Tag) error {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(getRegion())})
	if err != nil {
		return fmt.Errorf("Error getting AWS session while creating ebs tags: %v", err)
	}
	ec2Client := ec2.New(sess)

	input := &ec2.CreateTagsInput{
		Resources: []*string{volume},
		Tags:      tags,
	}
	_, err = ec2Client.CreateTags(input)
	if err != nil {
		return fmt.Errorf("failed to create ebs tags: %v", err)
	}
	return nil
}

func getRegion() string {
	region, present := os.LookupEnv("AWS_REGION")
	if !present {
		sess, err := session.NewSession()
		if err != nil {
			klog.Errorf("Error getting AWS session while retrieving region: %v", err)
		} else {
			svc := ec2metadata.New(sess)
			if r, err := svc.Region(); err == nil {
				region = r
			}
		}
	}
	return region
}
