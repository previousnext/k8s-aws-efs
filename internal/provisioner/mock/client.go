package mock

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/efs/efsiface"
)

// Client which mocks the EFS client.
type Client struct {
	efsiface.EFSAPI
	filesystems map[string]FileSystem
}

// FileSystem used for in memory mock storage.
type FileSystem struct {
	ID          string
	Tags        []Tag
	Performance string
	Mounts      []Mount
}

// Tag used for in memory mock storage.
type Tag struct {
	Key   string
	Value string
}

// Mount used for in memory mock storage.
type Mount struct {
	SubnetID string
}

// New mock EFS client.
func New() *Client {
	return &Client{
		filesystems: make(map[string]FileSystem),
	}
}

// DescribeFileSystems mock.
func (m *Client) DescribeFileSystems(input *efs.DescribeFileSystemsInput) (*efs.DescribeFileSystemsOutput, error) {
	output := &efs.DescribeFileSystemsOutput{}

	if fs, ok := m.filesystems[*input.CreationToken]; ok {
		output.FileSystems = []*efs.FileSystemDescription{
			{
				FileSystemId:    aws.String(fs.ID),
				LifeCycleState:  aws.String(efs.LifeCycleStateAvailable),
				PerformanceMode: aws.String(fs.Performance),
			},
		}
	}

	return output, nil
}

// CreateFileSystem mock.
func (m *Client) CreateFileSystem(input *efs.CreateFileSystemInput) (*efs.FileSystemDescription, error) {
	output := &efs.FileSystemDescription{}

	m.filesystems[*input.CreationToken] = FileSystem{
		ID:          *input.CreationToken,
		Performance: *input.PerformanceMode,
	}

	output.FileSystemId = input.CreationToken

	return output, nil
}

// CreateTags mock.
func (m *Client) CreateTags(input *efs.CreateTagsInput) (*efs.CreateTagsOutput, error) {
	output := &efs.CreateTagsOutput{}

	if fs, ok := m.filesystems[*input.FileSystemId]; ok {
		for _, tag := range input.Tags {
			fs.Tags = append(fs.Tags, Tag{
				Key:   *tag.Key,
				Value: *tag.Value,
			})
		}

		m.filesystems[*input.FileSystemId] = fs

		return output, nil
	}

	return output, errors.New("not found")
}

// DescribeMountTargets mock.
func (m *Client) DescribeMountTargets(input *efs.DescribeMountTargetsInput) (*efs.DescribeMountTargetsOutput, error) {
	output := &efs.DescribeMountTargetsOutput{}

	if fs, ok := m.filesystems[*input.FileSystemId]; ok {
		for _, mount := range fs.Mounts {
			output.MountTargets = []*efs.MountTargetDescription{
				{
					SubnetId:       aws.String(mount.SubnetID),
					LifeCycleState: aws.String(efs.LifeCycleStateAvailable),
				},
			}
		}

		return output, nil
	}

	return output, errors.New("filesystem not found")
}

// CreateMountTarget mock.
func (m *Client) CreateMountTarget(input *efs.CreateMountTargetInput) (*efs.MountTargetDescription, error) {
	output := &efs.MountTargetDescription{}

	if fs, ok := m.filesystems[*input.FileSystemId]; ok {
		fs.Mounts = append(fs.Mounts, Mount{
			SubnetID: *input.SubnetId,
		})

		m.filesystems[*input.FileSystemId] = fs

		output.SubnetId = input.SubnetId
		output.LifeCycleState = aws.String(efs.LifeCycleStateCreating)

		return output, nil
	}

	return output, errors.New("filesystem not found")
}
