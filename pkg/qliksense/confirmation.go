package qliksense

import (
	"fmt"
	"log"
	"strings"
)

func AskForConfirmation(s string) bool {
	for {
		fmt.Printf("%s [y/n]: ", s)
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			log.Fatal(err)
		}

		if strings.EqualFold(strings.ToLower(response), "y") || strings.EqualFold(strings.ToLower(response), "yes") {
			return true
		} else if strings.EqualFold(strings.ToLower(response), "n") || strings.EqualFold(strings.ToLower(response), "n") {
			return false
		}
	}
}