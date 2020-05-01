package preflight

import (
	"fmt"
)

func (qp *QliksensePreflight) CheckStorageClass(namespace string, kubeConfigContents []byte, cleanup bool) error {
	fmt.Println("Hello From Preflight storageclass check")
	return nil
}
