package provisioner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"
)

func TestFormatName(t *testing.T) {
	foo := controller.ProvisionOptions{
		PVName: "bar",
		PVC: &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "foo",
				Name:      "baz",
			},
		},
	}

	name, err := formatName("{{ .PVC.ObjectMeta.Namespace }}-{{ .PVName }}", foo)
	assert.Nil(t, err)
	assert.Equal(t, "foo-bar", name)

	name, err = formatName("{{ .PVC.ObjectMeta.Namespace }}-{{ .PVC.ObjectMeta.Name }}", foo)
	assert.Nil(t, err)
	assert.Equal(t, "foo-baz", name)
}
