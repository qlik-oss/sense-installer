package postflight

import (
	"fmt"

	. "github.com/logrusorgru/aurora"
	ansi "github.com/mattn/go-colorable"
	"github.com/pkg/errors"
)

func (qp *QliksensePostflight) RunAllPostflightChecks(namespace string, kubeConfigContents []byte, preflightOpts *PostflightOptions) error {
	checkCount := 0
	totalCount := 0

	out := ansi.NewColorableStdout()
	// Postflight db migration check
	if err := qp.DbMigrationCheck(namespace, kubeConfigContents); err != nil {
		fmt.Fprintf(out, "%s\n", Red("FAILED"))
		fmt.Printf("Error: %v\n\n", err)
	} else {
		fmt.Fprintf(out, "%s\n\n", Green("PASSED"))
		checkCount++
	}
	totalCount++

	if checkCount == totalCount {
		// All postflight checks were successful
		return nil
	}
	return errors.New("1 or more postflight checks have FAILED")
}
