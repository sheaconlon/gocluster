package gocluster

import (
	"testing"
	"io/ioutil"
	"bytes"
	"time"
	"strconv"
	"os"
)

func TestCommunication(t *testing.T) {
	file, err := ioutil.TempFile("", "tmp")
	Check(err)
	file.Write([]byte("Hello, world!"))
	file.Close()
	os.Mkdir("results", 0777)
	go func() {
		addr := ReceiveFiles(":59385", "/xyz", "f_i_e_l_d", "results", 3)
		if addr != "127.0.0.1" {
			t.Fail()
		}
	}()
	time.Sleep(time.Second * 5)
	for i := 0; i < 3; i++ {
		SendFile("127.0.0.1", ":59385", "/xyz", "f_i_e_l_d", file.Name())
	}
	for i := 0; i < 3; i++ {
		receivedData, err := ioutil.ReadFile("results" + "/" + strconv.Itoa(i))
		Check(err)
		if bytes.Compare(receivedData, []byte("Hello, world!")) != 0 {
			t.Fail()
		}
	}
	os.RemoveAll("results")
}
