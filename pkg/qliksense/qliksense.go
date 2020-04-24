package qliksense

import (
	"github.com/markbates/pkger"
)

// Qliksense is the logic behind the qliksense client
type Qliksense struct {
	QliksenseHome string
	CrdPkger      string
}

// New qliksense client, initialized with useful defaults.
func New(qliksenseHome string) *Qliksense {
	qliksenseClient := &Qliksense{
		QliksenseHome: qliksenseHome,
		CrdPkger:      pkger.Include("/pkg/qliksense/crds"),
	}
	return qliksenseClient
}
