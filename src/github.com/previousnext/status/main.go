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

		updates := make(map[string]*client.Efs)

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

			fmt.Println(name, "| Checking filesystem")

			// We start off by assuming that the filesystem is not ready.
			// That way we can fail early and be sure that we
			updates[name] = &client.Efs{
				Metadata: e.Metadata,
				Status: client.EfsStatus{
					LastUpdate:     time.Now(),
					LifeCycleState: client.LifeCycleStateUnknown,
				},
			}

			id, fs, err := checkFilesystem(svc, name)
			if err != nil {
				fmt.Println(name, "| Failed to check filesystem:", name, err)
				continue
			}

			// Add the ID to the list, that way the user can get a mount point.
			updates[name].Status.ID = id

			mnts, err := checkMounts(svc, id)
			if err != nil {
				fmt.Println(name, "| Failed to check filesystem mounts:", name, err)
				continue
			}

			if fs && mnts {
				updates[name].Status.LifeCycleState = client.LifeCycleStateReady
			} else {
				updates[name].Status.LifeCycleState = client.LifeCycleStateNotReady
			}

		}

		// Now we go through and patch all the EFS API objects.
		for name, u := range updates {
			// Lookup the existing object so we can push a new copy of it.
			e, err := efsClient.Get(u.Metadata.Namespace, u.Metadata.Name)
			if err != nil {
				fmt.Println(name, "| Failed to get existing filesystem for status update:", err)
				continue
			}

			// Override the current status with the new one.
			e.Status = u.Status

			// @todo, We need to figure out a way to do a Patch() vs Put().
			err = efsClient.Put(e)
			if err != nil {
				fmt.Println(name, "| Failed to update filesystem status:", err)
				continue
			}

			fmt.Println(name, "| Updated status:", u.Status.LifeCycleState)
		}
	}
}

// Helper function to check the status of a filesystem.
func checkFilesystem(svc *efs.EFS, name string) (string, bool, error) {
	fs, err := svc.DescribeFileSystems(&efs.DescribeFileSystemsInput{
		CreationToken: aws.String(name),
	})
	if err != nil {
		return "", false, err
	}

	if len(fs.FileSystems) != 1 {
		return "", false, fmt.Errorf("Filesystem not found")
	}

	return *fs.FileSystems[0].FileSystemId, true, nil
}

// Helper function to check the status of a filesystems mount points.
func checkMounts(svc *efs.EFS, id string) (bool, error) {
	resp, err := svc.DescribeMountTargets(&efs.DescribeMountTargetsInput{
		FileSystemId: aws.String(id),
	})
	if err != nil {
		return false, err
	}

	// If we have 0 mount targets, its not ready.
	if len(resp.MountTargets) == 0 {
		return false, nil
	}

	for _, m := range resp.MountTargets {
		if *m.LifeCycleState != efs.LifeCycleStateAvailable {
			return false, nil
		}
	}

	return true, nil
}
