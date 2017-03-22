package main

import (
	"fmt"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/efs"
	"k8s.io/client-go/rest"

	"github.com/previousnext/client"
)

var (
	cliSync = kingpin.Flag("sync-period", "How often to sync AWS EFS to K8s objects").Default("60s").OverrideDefaultFromEnvar("SYNC_PERIOD").Duration()
)

func main() {
	kingpin.Parse()

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Installing...")

	efsClient, err := client.NewClient(config)
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting...")

	limiter := time.Tick(*cliSync)

	for {
		<-limiter

		efsList, err := efsClient.ListAll()
		if err != nil {
			fmt.Println("Failed to load EFS objects:", err)
			continue
		}

		for _, e := range efsList.Items {
			var (
				svc  = efs.New(session.New(&aws.Config{Region: aws.String(e.Spec.Region)}))
				name = e.CreationToken()
			)

			err := e.Validate()
			if err != nil {
				fmt.Println(name, "| Not a valid EFS object:", err)
				continue
			}

			// Turn on the defaults (if things in the Spec aren't set.)
			e.Defaults()

			fmt.Println(name, "| Checking filesystem")

			// Ensures that we have created a filesystem.
			fs, err := createFilesystem(svc, name, e.Spec.Performance)
			if err != nil {
				fmt.Println(name, "| Failed to create filesystem:", err)
				continue
			}

			// Check if the filesystem is ready for mountpoints.
			if *fs.LifeCycleState != efs.LifeCycleStateAvailable {
				fmt.Println(name, "| Skipping mountpoints, filesystem lifecycle is:", *fs.LifeCycleState)
				continue
			}

			// Ensures that we have mount points created.
			for _, subnet := range e.Spec.Subnets {
				fmt.Println(name, "| Checking mount for subnet:", subnet)

				err := createMount(svc, *fs.FileSystemId, subnet, e.Spec.SecurityGroup)
				if err != nil {
					fmt.Println(name, "| Failed to create filesystem mountpoint:", err)
					continue
				}
			}
		}
	}

}

// Helper function to create an EFS filesystem.
func createFilesystem(svc *efs.EFS, name string, performance client.PerformanceMode) (*efs.FileSystemDescription, error) {
	describe, err := svc.DescribeFileSystems(&efs.DescribeFileSystemsInput{
		CreationToken: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	// We have found the filesystem! Give this back to the provisioner.
	if len(describe.FileSystems) > 0 {
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

// Helper function to create EFS mount points.
func createMount(svc *efs.EFS, id, subnet, security string) error {
	// Check if a mount exists in this subnet.
	mnts, err := svc.DescribeMountTargets(&efs.DescribeMountTargetsInput{
		FileSystemId: aws.String(id),
	})
	if err != nil {
		return err
	}

	for _, m := range mnts.MountTargets {
		// Check if we have already setup a mount point on a specific subnet.
		if *m.SubnetId == subnet {
			return nil
		}
	}

	// Create one if it does not exist.
	_, err = svc.CreateMountTarget(&efs.CreateMountTargetInput{
		FileSystemId: aws.String(id),
		SubnetId:     aws.String(subnet),
		SecurityGroups: []*string{
			aws.String(security),
		},
	})

	return err
}
