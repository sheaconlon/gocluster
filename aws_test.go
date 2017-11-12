package gocluster

import "testing"

// Check that creating an EC2 session, running a few instances, and then terminating them causes no errors.
func TestLifecycleNoErrors(t *testing.T) {
	var ec2Session = NewEC2Session()
	RunInstances(&ec2Session, 3)
	TerminateInstances(&ec2Session)
}
