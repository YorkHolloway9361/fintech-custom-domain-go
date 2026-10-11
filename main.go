package main

import (
	"context"
	"fmt"
	"log"
	"os"
)

func main() {
	domain, callback := "example.com", "https://pay.example.com/hooks/domain"
	if len(os.Args) > 1 {
		domain = os.Args[1]
	}
	c, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}
	secret, err := c.Onboard(context.Background(), domain, callback)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("onboarding started for %s; webhook secret returned once (%d bytes)\n", domain, len(secret))
}
