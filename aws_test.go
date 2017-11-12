package gocluster

import "testing"

func TestLifecycleNoErrors(t *testing.T) {
	var ec2Session = NewEC2Session()
	RunInstances(&ec2Session, 3)
	TerminateInstances(&ec2Session)
}
