package qliksense

// Qliksense is the logic behind the qliksense client
type Qliksense struct {
	PorterExe     string
	QliksenseHome string
}

// New qliksense client, initialized with useful defaults.
func New(porterExe, qliksenseHome string) *Qliksense {
	return &Qliksense{
		PorterExe:     porterExe,
		QliksenseHome: qliksenseHome,
	}
}
