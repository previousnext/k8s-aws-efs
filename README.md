Kubernetes - Storage Class - AWS EFS
====================================

[![CircleCI](https://circleci.com/gh/previousnext/k8s-aws-efs.svg?style=svg)](https://circleci.com/gh/previousnext/k8s-aws-efs)

**Maintainer**: Nick Schuch

Kubernetes storage class for automatically provisioning AWS EFS volumes.

This project would not be possible without:

https://github.com/kubernetes-incubator/external-storage

**Why not _external-storage/aws/efs?_**

That project uses an existing EFS filesystem and mounts subfolders for each PersistentVolumeClaim.

This project provisions a new EFS filesystem for each PersistentVolumeClaim, giving us:

* Security - Not all stored on the one filesystem
* Reliability - Other applications don't shared the same IOPs budget as your mount

## Usage

**Deploy the provisioner**

First we need to deploy our provisioner, this component is responsible for:
 
* Interfacing with a PersistentVolumeClaim
* Provisioning the required AWS EFS storage
* Returning the information needed to mount the storage

To deploy, create a file called `provisioner.yaml` with the contents below and run:

```bash
kubectl create -f provisioner.yaml
```

```yaml
kind: Deployment
apiVersion: extensions/v1beta1
metadata:
  name: aws-efs-provisioner
  namespace: kube-system
spec:
  replicas: 1
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: aws-efs-provisioner
    spec:
      containers:
        - name: aws-efs-provisioner
          image: previousnext/k8s-aws-efs:2.0.0
          env:
            - name:  EFS_PERFORMANCE
              value: "generalPurpose"
            - name:  AWS_REGION
              value: "ap-southeast-2"
            - name:  AWS_SECURITY_GROUP
              value: "sg-xxxxxxxxx"
            - name:  AWS_SUBNETS
              value: "subnet-xxxxxx,subnet-xxxxxx"
```

**Register our provisioner as a Storage Class**

Now we are going to register our storage class, this is way for us to map an "identifer" to our provsioner.

In this example we are mapping `aws-efs-gp` to our `storage.skpr.io/aws-efs-generalPurpose` provisioner.

To deploy, create a file called `class.yaml` with the contents below and run:

```bash
kubectl create -f class.yaml
```

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1beta1
metadata:
  name: aws-efs-gp
provisioner: efs.aws.skpr.io/generalPurpose
```

**Create your first test PersistentVolumeClaim**

Now we are going to provision our first claim, this will create an object that tells our provisioner to create
us an EFS storage volume.

To deploy, create a file called `test.yaml` with the contents below and run:

```bash
kubectl create -f test.yaml
```

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: test
  annotations:
    volume.beta.kubernetes.io/storage-class: "aws-efs-gp"
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      # This is not used by the provisioner, but is required by the PVC.
      storage: 1Mi
```

Now you can inspect the status of the PVC being provisioned with:

```bash
$ kubectl get pvc
NAME             STATUS    VOLUME        CAPACITY   ACCESSMODES   STORAGECLASS   AGE
test             Bound     fs-f6e605cf   8E         RWX           aws-efs-gp     5m
```

_NOTE: It will take 5(ish) minutes to get to the below state._

## AWS Configuration

**IAM Role**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "elasticfilesystem:DescribeFileSystems",
        "elasticfilesystem:CreateFileSystem",
        "elasticfilesystem:CreateTags",
        "elasticfilesystem:DescribeMountTargets",
        "elasticfilesystem:CreateMountTarget"
      ],
      "Resource": "*"
    }
  ]
}
```

**Credentials**

Before using the tool, ensure that you've configured credentials. The best
way to configure credentials on a development machine is to use the
`~/.aws/credentials` file, which might look like:

```ini
[default]
aws_access_key_id = AKID1234567890
aws_secret_access_key = MY-SECRET-KEY
```

You can learn more about the credentials file from this
[blog post](http://blogs.aws.amazon.com/security/post/Tx3D6U6WSFGOK2H/A-New-and-Standardized-Way-to-Manage-Credentials-in-the-AWS-SDKs).

Alternatively, you can set the following environment variables:

```
AWS_ACCESS_KEY_ID=AKID1234567890
AWS_SECRET_ACCESS_KEY=MY-SECRET-KEY
```

## Resources

* [Dynamic Provisioning and Storage Classes in Kubernetes](http://blog.kubernetes.io/2017/03/dynamic-provisioning-and-storage-classes-kubernetes.html)
* [Kubernetes Incubator: External Storage](https://github.com/kubernetes-incubator/external-storage)
* [Dave Cheney - Reproducible Builds](https://www.youtube.com/watch?v=c3dW80eO88I)

## Development

### Principles

* Code lives in the `workspace` directory

### Tools

* **Dependency management** - https://getgb.io
* **Build** - https://github.com/mitchellh/gox
* **Linting** - https://github.com/golang/lint

### Workflow

(While in the `workspace` directory)

**Installing a new dependency**

```bash
gb vendor fetch github.com/foo/bar
```

**Running quality checks**

```bash
make lint test
```

**Building binaries**

```bash
make build
```
