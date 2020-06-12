package qliksense

import (
	"fmt"
	"strings"

	"reflect"

	kconfig "github.com/qlik-oss/k-apis/pkg/config"
	"github.com/qlik-oss/sense-installer/pkg/api"
)

func (q *Qliksense) UnsetCmd(args []string) error {
	return unsetAll(q.QliksenseHome, args)
}
func unsetAll(qHome string, args []string) error {
	qConfig := api.NewQConfig(qHome)

	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		return err
	}

	// either delete all args or none
	for _, arg := range args {
		isRemoved := false
		// delete if key present
		if !strings.Contains(arg, ".") {
			if isRemoved = unsetOnlyKey(arg, qcr); isRemoved {
				//continue to the next arg
				continue
			} else if isRemoved = unsetServiceName(arg, qcr); isRemoved {
				//continue to the next arg
				continue
			} else {
				return fmt.Errorf("%s not found in the context", arg)
			}
		}
		// delete key inside configs if present
		// delete key inside secrets if present
		if isRemoved = unsetServiceKey(arg, qcr); isRemoved {
			//return qConfig.WriteCR(qcr)
			continue
		}
		if isRemoved = unsetTopAttrKey(arg, qcr); !isRemoved {
			return fmt.Errorf("%s not found in the context", arg)
		}
	}

	return qConfig.WriteCR(qcr)
}
func unsetOnlyKey(key string, qcr *api.QliksenseCR) bool {

	v := reflect.ValueOf(qcr.Spec).Elem().FieldByName(strings.Title(key))
	if v.IsValid() && v.CanSet() {
		v.Set(reflect.Zero(v.Type()))
		return true
	}
	return false
}

func unsetServiceName(svc string, qcr *api.QliksenseCR) bool {
	if qcr.Spec.Configs != nil && qcr.Spec.Configs[svc] != nil {
		delete(qcr.Spec.Configs, svc)
		return true
	}

	if qcr.Spec.Secrets != nil && qcr.Spec.Secrets[svc] != nil {
		delete(qcr.Spec.Secrets, svc)
		return true
	}
	return false
}

func unsetServiceKey(svcKey string, qcr *api.QliksenseCR) bool {
	sk := strings.Split(svcKey, ".")
	svc := sk[0]
	key := sk[1]

	if qcr.Spec.Configs != nil && qcr.Spec.Configs[svc] != nil {
		index := findIndex(key, qcr.Spec.Configs[svc])
		if index > -1 {
			qcr.Spec.Configs[svc][index] = qcr.Spec.Configs[svc][len(qcr.Spec.Configs[svc])-1]
			qcr.Spec.Configs[svc] = qcr.Spec.Configs[svc][:len(qcr.Spec.Configs[svc])-1]
			if len(qcr.Spec.Configs[svc]) == 0 {
				delete(qcr.Spec.Configs, svc)
			}
			return true
		}
	}

	if qcr.Spec.Secrets != nil && qcr.Spec.Secrets[svc] != nil {
		index := findIndex(key, qcr.Spec.Secrets[svc])
		if index > -1 {
			qcr.Spec.Secrets[svc][index] = qcr.Spec.Secrets[svc][len(qcr.Spec.Secrets[svc])-1]
			qcr.Spec.Secrets[svc] = qcr.Spec.Secrets[svc][:len(qcr.Spec.Secrets[svc])-1]
			if len(qcr.Spec.Secrets[svc]) == 0 {
				delete(qcr.Spec.Secrets, svc)
			}
			return true
		}
	}
	return false
}

func unsetTopAttrKey(attKey string, qcr *api.QliksenseCR) bool {
	sk := strings.Split(attKey, ".")
	attStruct := sk[0]
	key := sk[1]
	attV := reflect.ValueOf(qcr.Spec).Elem().FieldByName(strings.Title(attStruct))
	if !attV.IsValid() || attV.IsZero() || attV.IsNil() {
		return false
	}
	v := attV.Elem().FieldByName(strings.Title(key))
	if v.IsValid() && v.CanSet() {
		v.Set(reflect.Zero(v.Type()))
		return true
	}
	return false
}
func findIndex(elem string, nvs kconfig.NameValues) int {
	for i, nv := range nvs {
		if nv.Name == elem {
			return i
		}
	}
	return -1
}
