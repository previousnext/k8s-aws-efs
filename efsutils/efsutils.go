package efsutils

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/golang/glog"
)

// Create is used to create a new Elastic Filesystem.
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
			glog.Infof("FileSystem %s is ready", *fs.FileSystemId)
			break
		}

		<-limiter
	}

	var wg sync.WaitGroup
	// ResultChan, ErrorChan
	resc, errc := make(chan string), make(chan error)

	for _, subnet := range subnets {
		wg.Add(1)

		// Create the mount targets.
		go func(sn string) {
			defer wg.Done()

			_, err := CreateMount(client, *fs.FileSystemId, sn, securityGroup)
			if err != nil {
				errc <- err
			} else {
				for {
					glog.Infof("Waiting for mount target to become ready: %s:%s", name, sn)

					// Passing this back to the create function means that it will check if
					// the mount target exists first.
					// So it is safe for us to rerun this function to get the latest status.
					target, err := CreateMount(client, *fs.FileSystemId, sn, securityGroup)
					if err != nil {
						msg := fmt.Sprintf("Failed to create filesystem: %s", err)
						glog.Error(msg)
						errc <- errors.New(msg)
						break
					}

					// The filesystem is ready!
					if *target.LifeCycleState == efs.LifeCycleStateAvailable {
						glog.Infof("Mount point in subnet %s is available", sn)
						resc <- fmt.Sprintf("[subnet: %s]", sn)
						break
					}

					<-limiter
				}
			}
		}(subnet)
	}

	// If any errors come through, exit early.
	for i := 0; i < len(subnets); i++ {
		select {
		case res := <-resc:
			fmt.Printf("Created mount point %s\n", res)
		case err := <-errc:
			glog.Errorf("Error creating mount point %s", err)
			return "", err
		}
	}

	// Wait for all subnets to be ready
	wg.Wait()

	close(resc)
	close(errc)

	return *fs.FileSystemId, nil
}
