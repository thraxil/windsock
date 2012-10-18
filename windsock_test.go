package main

import (
	"fmt"
	"testing"
	"time"
)

type DummyAddr struct {
	S string
}

func (d DummyAddr) String() string {
	return d.S
}

func (d DummyAddr) Network() string {
	return d.S
}

func Test_validateToken(t *testing.T) {
	token := "anp8:1344361884:667494:127.0.0.1:306233f64522f1f970fc62fb3cf2d7320c899851"
	timestamp := time.Unix(1344361884, 0)
	remote_ip := DummyAddr{"127.0.0.1"}
	uni, err := validateToken(token, timestamp, remote_ip)
	if err != nil {
		t.Error("error validating")
	}
	if uni != "anp8" {
		fmt.Println(uni)
		t.Error("wrong uni")
	}
}

func Test_invalidToken(t *testing.T) {
	token := "anp8:1344361884:667494:127.0.0.1:306233f64522f1f970fc62fb3cf2d7320c899851GARBAGE"
	timestamp := time.Unix(1344361884, 0)
	remote_ip := DummyAddr{"127.0.0.1"}
	_, err := validateToken(token, timestamp, remote_ip)
	if err == nil {
		t.Error("invalid token. should've been an error here")
	}
}
