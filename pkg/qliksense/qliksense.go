package qliksense

// Qliksense is the logic behind the qliksense client
type Qliksense struct {
	QliksenseHome string
}

// New qliksense client, initialized with useful defaults.
func New(qliksenseHome string) *Qliksense {
	qliksenseClient := &Qliksense{
		QliksenseHome: qliksenseHome,
	}

	return qliksenseClient
}
