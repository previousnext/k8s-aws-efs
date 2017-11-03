package efsutils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
)

// Helper function to check if a filesystem exists, or create it if not present.
func CreateFilesystem(svc *efs.EFS, name string, performance string) (*efs.FileSystemDescription, error) {
	describe, err := svc.DescribeFileSystems(&efs.DescribeFileSystemsInput{
		CreationToken: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	// We have found the filesystem! Give this back to the provisioner.
	if len(describe.FileSystems) == 1 {
		return describe.FileSystems[0], nil
	}

	// We dont hav the filesystem, lets provision it now.
	create, err := svc.CreateFileSystem(&efs.CreateFileSystemInput{
		CreationToken:   aws.String(name),
		PerformanceMode: aws.String(string(performance)),
	})
	if err != nil {
		return nil, err
	}

	// Add tags to the filesystem, this makes it easier for site admins
	// to see what a filesystem was provisioned for.
	_, err = svc.CreateTags(&efs.CreateTagsInput{
		FileSystemId: create.FileSystemId,
		Tags: []*efs.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
		},
	})

	return create, nil
}
