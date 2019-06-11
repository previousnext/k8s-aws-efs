package provisioner

import (
	"bytes"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/efs/efsiface"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"
)

// Helper function for building hostname.
func formatName(format string, options controller.ProvisionOptions) (string, error) {
	var formatted bytes.Buffer

	t := template.Must(template.New("name").Parse(format))

	err := t.Execute(&formatted, options)
	if err != nil {
		return "", err
	}

	return formatted.String(), nil
}

// Helper function to check if a filesystem exists before creating.
func putFilesystem(svc efsiface.EFSAPI, name string, performance string) (*efs.FileSystemDescription, error) {
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

// Helper function to check if a mount exists before creating.
func putMount(svc efsiface.EFSAPI, id, subnet, security string) (*efs.MountTargetDescription, error) {
	// Check if a mount exists in this subnet.
	mnts, err := svc.DescribeMountTargets(&efs.DescribeMountTargetsInput{
		FileSystemId: aws.String(id),
	})
	if err != nil {
		return nil, err
	}

	for _, mount := range mnts.MountTargets {
		// Check if we have already setup a mount point on a specific subnet.
		if *mount.SubnetId == subnet {
			return mount, nil
		}
	}

	// Create one if it does not exist.
	return svc.CreateMountTarget(&efs.CreateMountTargetInput{
		FileSystemId: aws.String(id),
		SubnetId:     aws.String(subnet),
		SecurityGroups: []*string{
			aws.String(security),
		},
	})
}
