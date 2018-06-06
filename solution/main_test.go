package main

import (
	"errors"
	"fmt"
	"testing"
)

//Test function IsSyntacticallyValid
func TestIsSyntacticallyValid(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		outURL  string
		wantErr bool
	}{
		{"Pass wrong outURL", "http://foobar.com/fibo", "http://localhost:8090/fib", true},
		{"Pass invalid url", "http:foobar.com/fibo", "http://localhost:8090/fibo", true},
		{"Check HTTPS URL", "https://foobar.com/fibo", "https://localhost:8090/fibo", false},
		{"Check HTTP URL with fibo", "http://foobar.com/fibo", "http://localhost:8090/fibo", false},
		{"Check HTTP URL with primes", "http://example.com/primes", "http://localhost:8090/primes", false},
		{"Check HTTP URL with odd", "http://myserver.com/odd", "http://localhost:8090/odd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if u, ok := IsSyntacticallyValid(tt.url); (!ok || u != tt.outURL) && !tt.wantErr {
				fmt.Println(ok, u)
				var errMessage string
				if !ok {
					errMessage = "URL is not valid"
				} else if u != tt.outURL {
					errMessage = "url generated is not equal"
				}
				err := errors.New(errMessage)
				t.Errorf("IsSyntacticallyValid() error = %v", err.Error())
				return
			}
		})
	}

}
