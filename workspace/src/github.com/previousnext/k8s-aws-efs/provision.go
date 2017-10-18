package main

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/previousnext/efsutils"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MountOptionAnnotation is the annotation on a PV object that specifies a
// comma separated list of mount options
const MountOptionAnnotation = "volume.beta.kubernetes.io/mount-options"

// Provision creates a storage asset and returns a PV object representing it.
func (p *efsProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	// This is a consistent naming pattern for provisioning our EFS objects.
	name, err := formatName(p.format, options)
	if err != nil {
		return nil, err
	}

	glog.Infof("Provisioning filesystem: %s", name)

	id, err := efsutils.Create(p.region, name, p.subnets, p.securityGroup, p.performance)
	if err != nil {
		return nil, fmt.Errorf("failed to provision filesystem: %s", err)
	}

	glog.Infof("Responding with persistent volume spec: %s", name)

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
			Annotations: map[string]string{
				// https://kubernetes.io/docs/concepts/storage/persistent-volumes
				// http://docs.aws.amazon.com/efs/latest/ug/mounting-fs-mount-cmd-dns-name.html
				MountOptionAnnotation: "nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2",
			},
		},
		Spec: v1.PersistentVolumeSpec{
			// PersistentVolumeReclaimPolicy, AccessModes and Capacity are required fields.
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				// AWS EFS returns a "massive" file storage size when mounted. We replicate that here.
				v1.ResourceName(v1.ResourceStorage): resource.MustParse("8.0E"),
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				NFS: &v1.NFSVolumeSource{
					Server: fmt.Sprintf("%s.efs.%s.amazonaws.com", id, p.region),
					Path:   "/",
				},
			},
		},
	}

	return pv, nil
}

// Helper function for building hostname.
func formatName(format string, options controller.VolumeOptions) (string, error) {
	var formatted bytes.Buffer

	t := template.Must(template.New("name").Parse(format))

	err := t.Execute(&formatted, options)
	if err != nil {
		return "", err
	}

	return formatted.String(), nil
}
