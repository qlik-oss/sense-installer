package preflight

import (
	"github.com/qlik-oss/sense-installer/pkg/api"
	"github.com/qlik-oss/sense-installer/pkg/qliksense"
)

type PreflightOptions struct {
	Verbose      bool
	MongoOptions *MongoOptions
}

// // LogVerboseMessage logs a verbose message
// func (p *PreflightOptions) LogVerboseMessage(strMessage string, args ...interface{}) {
// 	if p.Verbose || os.Getenv("QLIKSENSE_DEBUG") == "true" {
// 		fmt.Printf(strMessage, args...)
// 	}
// }

type MongoOptions struct {
	MongodbUrl string
	CaCertFile string
}

type QliksensePreflight struct {
	Q  *qliksense.Qliksense
	P  *PreflightOptions
	CG *api.ClientGoUtils
}

func (qp *QliksensePreflight) GetPreflightConfigObj() *api.PreflightConfig {
	return api.NewPreflightConfig(qp.Q.QliksenseHome)
}

func (qp *QliksensePreflight) Cleanup(namespace string, kubeConfigContents []byte) error {
	qp.CG.LogVerboseMessage("Preflight clean\n")
	qp.CG.LogVerboseMessage("----------------\n")

	qp.CG.LogVerboseMessage("Removing deployment...\n")
	qp.CheckDeployment(namespace, kubeConfigContents, true)
	qp.CG.LogVerboseMessage("Removing service...\n")
	qp.CheckService(namespace, kubeConfigContents, true)
	qp.CG.LogVerboseMessage("Removing pod...\n")
	qp.CheckPod(namespace, kubeConfigContents, true)

	qp.CG.LogVerboseMessage("Removing role...\n")
	qp.CheckCreateRole(namespace, true)
	qp.CG.LogVerboseMessage("Removing rolebinding...\n")
	qp.CheckCreateRoleBinding(namespace, true)
	qp.CG.LogVerboseMessage("Removing serviceaccount...\n")
	qp.CheckCreateServiceAccount(namespace, true)

	qp.CG.LogVerboseMessage("Removing DNS check components...\n")
	qp.CheckDns(namespace, kubeConfigContents, true)
	qp.CG.LogVerboseMessage("Removing mongo check components...\n")
	qp.CheckMongo(kubeConfigContents, namespace, &PreflightOptions{MongoOptions: &MongoOptions{}}, true)
	return nil
}
