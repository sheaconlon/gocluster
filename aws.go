package gocluster

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"time"
)

// The name of the region for the EC2 instances
const Region string = "us-east-1"

// The ID of the image for the EC2 instances. This image must run Worker(...) on startup.
const ImageId string = "ami-130bdd69"

// The name of the type for the EC2 instances.
const InstanceType string = "t2.micro"

// The name of the keypair for the EC2 instances.
const KeyName string = "autodeploy-keypair"

// The ID of the subnet for the EC2 instances. This subnet must be set to auto-assign public IP addresses.
const SubnetId string = "subnet-c8221ac4"

// The duration to wait between successive polls of the EC2 instances' statuses, when waiting for them to enter the
// "running" state.
const runPollInterval time.Duration = 5 * time.Second

// An EC2 session.
type EC2Session struct {
	EC2Service * ec2.EC2
	Reservations []*ec2.Reservation
}

// NewEC2Session creates a new EC2 session. Your ~/.aws/credentials file must contain the following for your root AWS
// account or an appropriate IAM user. Otherwise, this will fail.
// [default]
// aws_access_key_id = <YOUR_ACCESS_KEY_ID>
// aws_secret_access_key = <YOUR_SECRET_ACCESS_KEY>
func NewEC2Session() (ec2Session EC2Session) {
	ec2Session = EC2Session{}
	awsSession, err := session.NewSession(&aws.Config{Region: aws.String(Region)})
	Check(err)
	ec2Session.EC2Service = ec2.New(awsSession)
	return
}

// RunInstances launches some EC2 instances and returns their public IP addresses.
func RunInstances(ec2Session *EC2Session, count int) (addresses []string) {
	// Request the instances.
	var ImageId string = ImageId
	var InstanceType string = InstanceType
	var KeyName string = KeyName
	var Count int64 = int64(count)
	var SubnetId string = SubnetId
	input := ec2.RunInstancesInput{
		ImageId: &ImageId,
		InstanceType: &InstanceType,
		KeyName: &KeyName,
		MaxCount: &Count,
		MinCount: &Count,
		SubnetId: &SubnetId,
	}
	reservation, err := ec2Session.EC2Service.RunInstances(&input)
	Check(err)
	ec2Session.Reservations = append(ec2Session.Reservations, reservation)

	// Wait for the instances to enter the "running" state.
	var ids []*string
	for _, instance := range reservation.Instances {
		ids = append(ids, instance.InstanceId)
	}
	ticker := time.NewTicker(runPollInterval)
	for _ = range ticker.C {
		var allReady bool = true
		input := ec2.DescribeInstancesInput{InstanceIds : ids}
		output, err := ec2Session.EC2Service.DescribeInstances(&input)
		Check(err)
		// NOTE: Assumes only one page of results.
		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				if *instance.State.Name == "running" {
					addresses = append(addresses, *instance.PublicIpAddress)
				} else {
					allReady = false
				}
			}
		}
		if allReady {
			ticker.Stop()
			break
		}
	}

	return
}

// TerminateInstances terminates all the EC2 instances that have been run using some session.
func TerminateInstances(ec2Session * EC2Session) {
	var InstanceIds []*string
	for _, reservation := range ec2Session.Reservations {
		for _, instance := range reservation.Instances {
			InstanceIds = append(InstanceIds, instance.InstanceId)
		}
	}
	input := ec2.TerminateInstancesInput{
		InstanceIds: InstanceIds,
	}
	_, err := ec2Session.EC2Service.TerminateInstances(&input)
	Check(err)
}