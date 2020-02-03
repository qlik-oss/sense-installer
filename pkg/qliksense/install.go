package qliksense

import (
	"fmt"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) InstallQK8s(version string) {

	// step1: fetch 1.0.0 # pull down qliksense-k8s@1.0.0
	// step2: operator view | kubectl apply -f # operator manifest (CRD)
	// step3: config apply | kubectl apply -f # generates patches (if required) in configuration directory, applies manifest
	// step4: config view | kubectl apply -f # generates Custom Resource manifest (CR)

	//io.WriteString(os.Stdout, q.GetCRDString())
	//fmt.Println(version)
	fmt.Println("Fetching " + version)
	//qConfig := qapi.NewQConfig(q.QliksenseHome)
	//qcr, err := qConfig.GetCurrentCR()

	//TODO: may need to check if CRD already installed, but doing apply does not hurt for now
	//install crd into cluster
	qapi.KubectlApply(q.GetCRDString())
}
