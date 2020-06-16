package qliksense

func (q *Qliksense) ApplyCRFromBytes(crBytes []byte, opts *InstallCommandOptions, overwriteExistingContext bool) error {
	if err := q.LoadCr(crBytes, overwriteExistingContext); err != nil {
		return err
	}
	return q.InstallQK8s("", opts)
}
