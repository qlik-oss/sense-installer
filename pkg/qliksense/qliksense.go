//go:generate packr2
package qliksense

import (
	"github.com/gobuffalo/packr/v2"
)

// Qliksense is the logic behind the qliksense client
type Qliksense struct {
	QliksenseHome string
	CrdBox        *packr.Box ``
}

// New qliksense client, initialized with useful defaults.
func New(qliksenseHome string) *Qliksense {
	qliksenseClient := &Qliksense{
		QliksenseHome: qliksenseHome,
		CrdBox:        packr.New("crds", "./crds"),
	}

	return qliksenseClient
}
