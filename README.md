Kubernetes - AWS EFS
====================

## Overview

The following project provides a method for:

* Provisioning AWS EFS resources
* Being a low level tool for higher level deployment frameworks (Internal project "Skipper")

The following project is broken up into the following components:

* **Provisioner** - Creates EFS resources
* **Status** - Updates Kubernetes with the status of the resources
* **CLI** - Used for listing all the provisioned resources and their status

![Architecture](/docs/diagram.png "Architecture")

## Usage

**Deploy the Kubernetes controller**

Create a contoller in Kubernetes with `kubectl create -f example/controller.yaml`

**Define a new EFS resource**

Create a resource in Kubernetes with `kubectl create -f example/efs.yaml`

**Viewing the status of EFS resources provisioned**

Using the CLI binary providied in the Docker image.

(Get a shell in the running pod)

```bash
$ cli

NAMESPACE       NAME            REGION          SUBNETS                                 SECURITY        ID              CYCLE   LAST UPDATE
test            testing1        ap-southeast-2  subnet-XXXXXXXX, subnet-XXXXXXXX        sg-XXXXXXXX     fs-XXXXXXXX     Ready   2017-03-22 19:27:42.820665151 +1000 AEST
test            testing2        ap-southeast-2  subnet-XXXXXXXX, subnet-XXXXXXXX        sg-XXXXXXXX     fs-XXXXXXXX     Ready   2017-03-22 19:27:42.965406845 +1000 AEST
```

## IAM

Permissions required for the controller components:

* efs:DescribeFileSystems
* efs:CreateFileSystem
* efs:CreateTags
* efs:DescribeMountTargets
* efs:CreateMountTarget
