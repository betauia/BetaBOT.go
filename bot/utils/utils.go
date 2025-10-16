package utils

import "log"

func CheckNilErr(err error) {
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}
