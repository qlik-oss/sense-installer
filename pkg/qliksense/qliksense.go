//go:generate packr2
package qliksense

import (
	"os"
	"path"

	"github.com/gobuffalo/packr/v2"
)

// Qliksense is the logic behind the qliksense client
type Qliksense struct {
	QliksenseHome        string
	QliksenseEjsonKeyDir string
	CrdBox               *packr.Box ``
}

// New qliksense client, initialized with useful defaults.
func New(qliksenseHome string) (*Qliksense, error) {
	qliksenseClient := &Qliksense{
		QliksenseHome: qliksenseHome,
		CrdBox:        packr.New("crds", "./crds"),
	}

	qliksenseClient.QliksenseEjsonKeyDir = path.Join(qliksenseHome, "ejson", "keys")
	if err := os.MkdirAll(qliksenseClient.QliksenseEjsonKeyDir, os.ModePerm); err != nil {
		return nil, err
	}
	return qliksenseClient, nil
}
