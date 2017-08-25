package efsutils

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/golang/glog"
)

func Create(region, name string, subnets []string, securityGroup string, performance string) (string, error) {
	var (
		client = efs.New(session.New(&aws.Config{Region: aws.String(region)}))

		// Limiter used to wait for the a filesystem to get created.
		limiter = time.Tick(time.Second * 15)
	)

	// Ensures that we have created a filesystem.
	fs, err := CreateFilesystem(client, name, performance)
	if err != nil {
		return "", fmt.Errorf("failed to create filesystem: %s", err)
	}

	// Wait for the filesystem to become available.
	// @todo, We need to have a "bail out" point just to be sure we are locked in a loop.
	for {
		glog.Infof("Waiting for filesystem to become ready: %s", name)

		// Passing this back to the create function means that it will check if the filesystem exists first.
		// So it is safe for us to rerun this function to get the latest status.
		fs, err := CreateFilesystem(client, name, performance)
		if err != nil {
			return "", fmt.Errorf("failed to create filesystem: %s", err)
		}

		// The filesystem is ready!
		if *fs.LifeCycleState == efs.LifeCycleStateAvailable {
			break
		}

		<-limiter
	}

	// Create the mount targets.
	// @todo, Make this run in parrallel for speed.
	for _, subnet := range subnets {
		_, err := CreateMount(client, *fs.FileSystemId, subnet, securityGroup)
		if err != nil {
			return "", err
		}

		for {
			glog.Infof("Waiting for mount target to become ready: %s", name)

			// Passing this back to the create function means that it will check if the mount target exists first.
			// So it is safe for us to rerun this function to get the latest status.
			target, err := CreateMount(client, *fs.FileSystemId, subnet, securityGroup)
			if err != nil {
				return "", fmt.Errorf("failed to create filesystem: %s", err)
			}

			// The filesystem is ready!
			if *target.LifeCycleState == efs.LifeCycleStateAvailable {
				break
			}

			<-limiter
		}
	}

	return *fs.FileSystemId, nil
}
