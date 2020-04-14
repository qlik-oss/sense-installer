package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-tty"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

type eulaPreRunHooksT struct {
	validators              map[string]func(cmd *cobra.Command, q *qliksense.Qliksense) (bool, error)
	postValidationArtifacts map[string]interface{}
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

func (e *eulaPreRunHooksT) addPostValidationArtifact(artifactName string, artifact interface{}) {
	e.postValidationArtifacts[artifactName] = artifact
}

func (e *eulaPreRunHooksT) getPostValidationArtifact(artifactName string) interface{} {
	if artifact, ok := e.postValidationArtifacts[artifactName]; ok {
		return artifact
	}
	return nil
}

var eulaEnforced = os.Getenv("QLIKSENSE_EULA_ENFORCE") == "true"
var eulaText = "Please read the end user license agreement at: https://www.qlik.com/us/legal/license-terms"
var eulaPrompt = "Do you accept our EULA? (y/n): "
var eulaErrorInstruction = `You must enter "y" to continue`
var eulaPreRunHooks = eulaPreRunHooksT{
	validators:              make(map[string]func(cmd *cobra.Command, q *qliksense.Qliksense) (bool, error)),
	postValidationArtifacts: make(map[string]interface{}),
}

func commandAlwaysRequiresEulaAcceptance(commandName string) bool {
	return commandName == "install" || commandName == "upgrade" || commandName == "apply"
}

func globalEulaPreRun(cmd *cobra.Command, q *qliksense.Qliksense) {
	if isEulaEnforced(cmd.CommandPath()) {
		if strings.TrimSpace(strings.ToLower(cmd.Flag("acceptEULA").Value.String())) != "yes" {
			if eulaPreRunHook := eulaPreRunHooks.getValidator(cmd.CommandPath()); eulaPreRunHook != nil {
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
}

func globalEulaPostRun(cmd *cobra.Command, q *qliksense.Qliksense) {
	if isEulaEnforced(cmd.CommandPath()) {
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
	answer := readRuneFromTty()
	if strings.ToLower(answer) != "y" {
		fmt.Println(eulaErrorInstruction)
		os.Exit(1)
	}
}

func readRuneFromTty() string {
	t, err := tty.Open()
	if err != nil {
		panic(err)
	}
	defer t.Close()
	answer, err := t.ReadString()
	if err != nil {
		panic(err)
	}
	return answer
}
