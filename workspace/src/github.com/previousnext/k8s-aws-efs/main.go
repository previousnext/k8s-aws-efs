package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	// Create the provisioner: it implements the Provisioner interface expected by the controller.
	apiVersion, provisioner, err := NewProvisioner()
	if err != nil {
		glog.Fatal("Failed to create provisioner: %s", err)
	}

	glog.Infof("Running provisioner: %s", apiVersion)

	// Start the provision controller which will dynamically provision NFS PVs
	pc := controller.NewProvisionController(clientset, apiVersion, provisioner, serverVersion.GitVersion, controller.CreateProvisionedPVInterval(time.Minute*10), controller.LeaseDuration(time.Minute*10))
	pc.Run(wait.NeverStop)
}

// NewProvisioner is used to build an EFS provisioner.
func NewProvisioner() (string, controller.Provisioner, error) {
	// http://docs.aws.amazon.com/efs/latest/ug/performance.html#performancemodes
	performance := os.Getenv("EFS_PERFORMANCE")
	if performance == "" {
		performance = efs.PerformanceModeGeneralPurpose
	}

	apiVersion := os.Getenv("API_VERSION")
	if apiVersion == "" {
		// We use the "performance" type as part of the apiVersion. This allows us to have a provisioner for both
		// types of storage eg.
		//   * skpr.io/aws/efs/generalPurpose
		//   * skpr.io/aws/efs/maxIO
		apiVersion = fmt.Sprintf("efs.aws.skpr.io/%s", performance)
	}

	// Region to provision the storage in.
	region := os.Getenv("AWS_REGION")
	if region == "" {
		return "", nil, fmt.Errorf("environment variable AWS_REGION not found")
	}

	// AWS VPC Subnets to deploy the EFS "mount points" to.
	// http://docs.aws.amazon.com/efs/latest/ug/accessing-fs.html
	subnets := os.Getenv("AWS_SUBNETS")
	if subnets == "" {
		return "", nil, fmt.Errorf("environment variable AWS_SUBNETS not found")
	}

	// AWS_SECURITY_GROUP assigns a VPC security group to the mount points.
	// http://docs.aws.amazon.com/efs/latest/ug/accessing-fs.html
	security := os.Getenv("AWS_SECURITY_GROUP")
	if security == "" {
		return "", nil, fmt.Errorf("environment variable AWS_SECURITY_GROUP not found")
	}

	// EFS_NAME_FORMAT allows for backwards compatibility with other EFS tools.
	//   eg. My existing EFS resources use the pattern "{{ .PVC.ObjectMeta.Namespace }}-{{ .PVC.ObjectMeta.Name }}"
	format := os.Getenv("EFS_NAME_FORMAT")
	if format == "" {
		format = "{{ .PVC.ObjectMeta.Namespace }}-{{ .PVName }}"
	}

	provisioner := &efsProvisioner{
		region:        region,
		securityGroup: security,
		subnets:       strings.Split(subnets, ","),
		performance:   performance,
		format:        format,
	}

	return apiVersion, provisioner, nil
}
