package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

type eulaPreRunHooksT struct {
	validators              map[string]func(cmd *cobra.Command, q *qliksense.Qliksense) (bool, error)
	postValidationArtifacts map[string]map[string]interface{}
}

func (e *eulaPreRunHooksT) addValidator(command string, validator func(cmd *cobra.Command, q *qliksense.Qliksense) (bool, error)) {
	e.validators[command] = validator
}

func (e *eulaPreRunHooksT) getValidator(command string) func(cmd *cobra.Command, q *qliksense.Qliksense) (bool, error) {
	if validator, ok := e.validators[command]; ok {
		return validator
	}
	return nil
}

func (e *eulaPreRunHooksT) addPostValidationArtifact(command string, artifactName string, artifact interface{}) {
	if _, ok := e.postValidationArtifacts[command]; !ok {
		e.postValidationArtifacts[command] = make(map[string]interface{})
	}
	e.postValidationArtifacts[command][artifactName] = artifact
}

func (e *eulaPreRunHooksT) getPostValidationArtifact(command string, artifactName string) interface{} {
	if artifacts, ok1 := e.postValidationArtifacts[command]; ok1 {
		if artifact, ok2 := artifacts[artifactName]; ok2 {
			return artifact
		}
	}
	return nil
}

var eulaEnforced = os.Getenv("QLIKSENSE_EULA_ENFORCE") == "true"
var eulaText = "Please read the end user license agreement at: https://www.qlik.com/us/legal/license-terms"
var eulaPrompt = "Do you accept our EULA? (y/n): "
var eulaErrorInstruction = "You must enter y/yes to continue"
var eulaPreRunHooks = eulaPreRunHooksT{
	validators:              make(map[string]func(cmd *cobra.Command, q *qliksense.Qliksense) (bool, error)),
	postValidationArtifacts: make(map[string]map[string]interface{}),
}
var eulaAcceptedFromPrompt = false

func commandAlwaysRequiresEulaAcceptance(commandName string) bool {
	return commandName == "install" || commandName == "apply"
}

func globalEulaPreRun(cmd *cobra.Command, q *qliksense.Qliksense) {
	if isEulaEnforced(cmd.Name()) {
		if eulaPreRunHook := eulaPreRunHooks.getValidator(cmd.Name()); eulaPreRunHook != nil {
			if eulaAccepted, err := eulaPreRunHook(cmd, q); err != nil {
				panic(err)
			} else if !eulaAccepted {
				doEnforceEula()
			}
		} else if qConfig, err := qapi.NewQConfigE(q.QliksenseHome); err != nil {
			doEnforceEula()
		} else if qcr, err := qConfig.GetCurrentCR(); err != nil || !qcr.IsEULA() {
			doEnforceEula()
		}
	}
}

func globalEulaPostRun(_ *cobra.Command, q *qliksense.Qliksense) {
	if eulaAcceptedFromPrompt {
		if err := q.SetEulaAccepted(); err != nil {
			panic(err)
		}
	}
}

func isEulaEnforced(commandName string) bool {
	return eulaEnforced || commandAlwaysRequiresEulaAcceptance(commandName)
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
	eulaAcceptedFromPrompt = true
}
