package qliksense

import (
	"fmt"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) InstallQK8s(version string) {
	//io.WriteString(os.Stdout, q.GetCRDString())
	fmt.Println(version)
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qConfig.GetCurrentCR()
}
