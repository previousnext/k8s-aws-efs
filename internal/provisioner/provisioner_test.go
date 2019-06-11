package provisioner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"

	"github.com/previousnext/k8s-aws-efs/internal/provisioner/mock"
)

func TestProvisioner(t *testing.T) {
	params := Params{
		Region:        "ap-southeast-2",
		Format:        "{{ .PVC.ObjectMeta.Namespace }}-{{ .PVName }}",
		Performance:   "generalPurpose",
		SecurityGroup: "sg-xxxxxxxxxxxx",
		Subnets: []string{
			"subnet-xxxxxxxx",
			"subnet-yyyyyyyy",
		},
	}

	provisioner, err := New(mock.New(), params)
	assert.Nil(t, err)

	options := controller.ProvisionOptions{
		PVName: "test",
		PVC: &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "namespace",
			},
		},
	}

	volume, err := provisioner.Provision(options)
	assert.Nil(t, err)

	want := corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "namespace-test",
			Annotations: map[string]string{
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
					Server: "namespace-test.efs.ap-southeast-2.amazonaws.com",
					Path:   "/",
				},
			},
		},
	}

	assert.Equal(t, want, *volume)

	err = provisioner.Delete(volume)
	assert.Nil(t, err)
}
