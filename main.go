package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/golang/glog"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"

	"github.com/previousnext/k8s-aws-efs/internal/provisioner"
)

func main() {
	flag.Parse()
	flag.Set("logtostderr", "true")

	// Create an InClusterConfig and use it to create a client for the controller
	// to use to communicate with Kubernetes
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal("Failed to create config: %s", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("Failed to create client: %s", err)
	}

	// The controller needs to know what the server version is because out-of-tree
	// provisioners aren't officially supported until 1.5
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		glog.Fatal("Error getting server version: %s", err)
	}

	var params provisioner.Params

	err = envconfig.Process("provisioner", &params)
	if err != nil {
		glog.Fatal("Failed to load params: %s", err)
	}

	apiVersion := os.Getenv("API_VERSION")
	if apiVersion == "" {
		// We use the "performance" type as part of the apiVersion. This allows us to have a provisioner for both
		// types of storage eg.
		//   * skpr.io/aws/efs/generalPurpose
		//   * skpr.io/aws/efs/maxIO
		apiVersion = fmt.Sprintf("efs.aws.skpr.io/%s", params.Performance)
	}

	client := efs.New(session.New())

	provisioner, err := provisioner.New(client, params)
	if err != nil {
		glog.Fatal("Failed to create provisioner: %s", err)
	}

	glog.Infof("Running provisioner: %s", apiVersion)

	// Start the provision controller which will dynamically provision NFS PVs
	pc := controller.NewProvisionController(clientset, apiVersion, provisioner, serverVersion.GitVersion, controller.CreateProvisionedPVInterval(time.Minute*10), controller.LeaseDuration(time.Minute*10))
	pc.Run(wait.NeverStop)
}
