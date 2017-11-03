package main

import (
	"testing"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFormatName(t *testing.T) {
	foo := controller.VolumeOptions{
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
