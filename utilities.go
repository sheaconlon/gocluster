package gocluster

import (
	"fmt"
	"os"
)

// Check an error. Print it and exit if not nil.
func Check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}