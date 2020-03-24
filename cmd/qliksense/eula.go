package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
)

var eulaEnforced = false
var eulaText = "EULA text goes here..."
var eulaPrompt = "Do you accept our EULA? (y/n): "
var eulaErrorInstruction = "You must enter y/yes to continue"

func isEulaEnforced() bool {
	return eulaEnforced
}

func enforceEula(q *qliksense.Qliksense) {
	if isEulaEnforced() {
		if qConfig, err := qapi.NewQConfigE(q.QliksenseHome); err != nil {
			doEnforceEula()
		} else if qcr, err := qConfig.GetCurrentCR(); err != nil || !qcr.IsEULA() {
			doEnforceEula()
		}
	}
}

func doEnforceEula() {
	fmt.Println(eulaText)
	fmt.Print(eulaPrompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanSuccess := scanner.Scan()
	if !scanSuccess {
		fmt.Println(eulaErrorInstruction)
		os.Exit(1)
	}
	line := scanner.Text()
	answer := strings.ToLower(strings.TrimSpace(line))
	if answer != "y" && answer != "yes" {
		fmt.Println(eulaErrorInstruction)
		os.Exit(1)
	}
}
