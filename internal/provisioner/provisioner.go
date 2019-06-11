package provisioner

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/efs/efsiface"
	"github.com/golang/glog"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"
)

var _ controller.Provisioner = &Provisioner{}

// Provisioner for creating volumes.
type Provisioner struct {
	client efsiface.EFSAPI
	params Params
}

// Params required for provisioning volumes.
type Params struct {
	Region        string   `envconfig:"AWS_REGION"         default:"ap-southeast-2"`
	Format        string   `envconfig:"EFS_NAME_FORMAT"    default:"{{ .PVC.ObjectMeta.Namespace }}-{{ .PVName }}"`
	Performance   string   `envconfig:"EFS_PERFORMANCE"    default:"generalPurpose"`
	SecurityGroup string   `envconfig:"AWS_SECURITY_GROUP" required:"true"`
	Subnets       []string `envconfig:"AWS_SUBNETS"        required:"true"`
}

// New provisioner for creating and deleting EFS volumes.
func New(client efsiface.EFSAPI, params Params) (controller.Provisioner, error) {
	provisioner := &Provisioner{
		client: client,
		params: params,
	}

	return provisioner, nil
}

// Provision creates a storage asset and returns a PV object representing it.
func (p *Provisioner) Provision(options controller.ProvisionOptions) (*corev1.PersistentVolume, error) {
	// This is a consistent naming pattern for provisioning our EFS objects.
	name, err := formatName(p.params.Format, options)
	if err != nil {
		return nil, err
	}

	glog.Infof("Provisioning filesystem: %s", name)

	// Limiter used to wait for the a filesystem to get created.
	limiter := time.Tick(time.Second * 15)

	// Ensures that we have created a filesystem.
	fs, err := putFilesystem(p.client, name, p.params.Performance)
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem: %s", err)
	}

	// Wait for the filesystem to become available.
	for {
		glog.Infof("Waiting for filesystem to become ready: %s", name)

		// Passing this back to the create function means that it will check if the filesystem exists first.
		// So it is safe for us to rerun this function to get the latest status.
		fs, err := putFilesystem(p.client, name, p.params.Performance)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem: %s", err)
		}

		// The filesystem is ready!
		if *fs.LifeCycleState == efs.LifeCycleStateAvailable {
			break
		}

		<-limiter
	}

	var group errgroup.Group

	// Create the mount targets.
	for _, subnet := range p.params.Subnets {
		group.Go(func() error {
			_, err := putMount(p.client, *fs.FileSystemId, subnet, p.params.SecurityGroup)
			if err != nil {
				return err
			}

			for {
				glog.Infof("Waiting for mount target to become ready: %s", name)

				// Passing this back to the create function means that it will check if the mount target exists first.
				// So it is safe for us to rerun this function to get the latest status.
				target, err := putMount(p.client, *fs.FileSystemId, subnet, p.params.SecurityGroup)
				if err != nil {
					return fmt.Errorf("failed to create filesystem: %s", err)
				}

				// The filesystem is ready!
				if *target.LifeCycleState == efs.LifeCycleStateAvailable {
					break
				}

				<-limiter
			}

			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to create mount: %s", err)
	}

	glog.Infof("Responding with persistent volume spec: %s", name)

	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: *fs.FileSystemId,
			Annotations: map[string]string{
				// https://kubernetes.io/docs/concepts/storage/persistent-volumes
				// http://docs.aws.amazon.com/efs/latest/ug/mounting-fs-mount-cmd-dns-name.html
				MountOptionAnnotation: "nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2",
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			// PersistentVolumeReclaimPolicy, AccessModes and Capacity are required fields.
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: corev1.ResourceList{
				// AWS EFS returns a "massive" file storage size when mounted. We replicate that here.
				corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("8.0E"),
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server: fmt.Sprintf("%s.efs.%s.amazonaws.com", *fs.FileSystemId, p.params.Region),
					Path:   "/",
				},
			},
		},
	}

	return pv, nil
}

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
// @todo, Tag FileSystem as "ready for removal"
// @todo, Tag FileSystem with a date to show how old it is.
func (p *Provisioner) Delete(volume *corev1.PersistentVolume) error {
	return nil
}
