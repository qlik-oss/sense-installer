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

		if strings.EqualFold(response, "y") || strings.EqualFold(response, "yes") {
			return true
		} else if strings.EqualFold(response, "n") || strings.EqualFold(response, "no") {
			return false
		}
	}
}
