package main

import (
	"github.com/kubernetes-incubator/external-storage/lib/controller"
)

var _ controller.Provisioner = &efsProvisioner{}

type efsProvisioner struct {
	// Region to provision the new EFS filesystem.
	region string

	// The AWS security group assigned to the EFS filesystem mount point.
	securityGroup string

	// Subnets to provision mount points.
	subnets []string

	// Performance.
	performance string

	// Formatting used to derive the name used in EFS.
	format string
}
