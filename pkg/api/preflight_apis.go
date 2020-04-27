package api

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PreflightConfig struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec              *PreflightSpec `json:"spec" yaml:"spec"`
	QliksenseHomePath string         `json:"-" yaml:"-"`
}

type PreflightSpec struct {
	MinK8sVersion   string            `json:"minK8sVersion,omitempty" yaml:"minK8sVersion,omitempty"`
	MinMongoVersion string            `json:"minMongoVersion,omitempty" yaml:"minMongoVersion,omitempty"`
	Images          map[string]string `json:"images,omitempty" yaml:"images,omitempty"`
}

//NewPreflightConfigEmpty create empty PreflightConfig object
func NewPreflightConfigEmpty(qHome string) *PreflightConfig {
	p := &PreflightConfig{
		QliksenseHomePath: qHome,
		TypeMeta: metav1.TypeMeta{
			APIVersion: "config.qlik.com/v1",
			Kind:       "PreflightConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "PreflightConfigMetadata",
		},
		Spec: &PreflightSpec{},
	}
	return p
}

//NewPreflightConfig create empty PreflightConfig object if preflit/preflight-config.yaml not exist
func NewPreflightConfig(qHome string) *PreflightConfig {
	p := NewPreflightConfigEmpty(qHome)
	conFile := p.GetConfigFilePath()
	if _, err := os.Lstat(conFile); err != nil {
		return p
	}
	p = &PreflightConfig{
		QliksenseHomePath: qHome,
	}
	if err := ReadFromFile(p, conFile); err != nil {
		return nil
	}
	return p
}

//GetConfigFilePath return preflight-config.yaml file path
func (p *PreflightConfig) GetConfigFilePath() string {
	return filepath.Join(p.QliksenseHomePath, "preflight", "preflight-config.yaml")
}

//Write write PreflightConfig object into the ~/.qliksense/preflight/preflight-config.yaml file
func (p *PreflightConfig) Write() error {
	pDir := filepath.Join(p.QliksenseHomePath, "preflight")
	if err := os.MkdirAll(pDir, os.ModePerm); err != nil {
		return err
	}
	return WriteToFile(p, p.GetConfigFilePath())
}

func (p *PreflightConfig) AddMinK8sV(version string) {
	if p.Spec == nil {
		p.Spec = &PreflightSpec{}
	}
	p.Spec.MinK8sVersion = version
}

func (p *PreflightConfig) AddMinMongoV(version string) {
	if p.Spec == nil {
		p.Spec = &PreflightSpec{}
	}
	p.Spec.MinMongoVersion = version
}

func (p *PreflightConfig) AddImage(imageFor, imageName string) {
	if p.Spec.Images == nil {
		p.Spec.Images = make(map[string]string)
	}
	p.Spec.Images[imageFor] = imageName
}

func (p *PreflightConfig) GetImageName(imageFor string, accountForImageRegistry bool) (string, error) {
	if p.Spec.Images == nil {
		return "", nil
	}
	image := p.Spec.Images[imageFor]
	if accountForImageRegistry {
		qConfig := NewQConfig(p.QliksenseHomePath)
		if currentCR, err := qConfig.GetCurrentCR(); err != nil {
			return "", err
		} else if imageRegistry := currentCR.Spec.GetImageRegistry(); imageRegistry != "" {
			imageSegments := strings.Split(image, "/")
			imageNameAndTag := imageSegments[len(imageSegments)-1]
			return path.Join(imageRegistry, imageNameAndTag), nil
		}
	}
	return image, nil
}
func (p *PreflightConfig) GetMinK8sVersion() string {
	return p.Spec.MinK8sVersion
}

func (p *PreflightConfig) GetMinMongoVersion() string {
	return p.Spec.MinMongoVersion
}

func (p *PreflightConfig) IsExistOnDisk() bool {
	if _, err := os.Lstat(p.GetConfigFilePath()); err != nil {
		return false
	}
	return true
}

func (p *PreflightConfig) GetImageMap() map[string]string {
	return p.Spec.Images
}

func (p *PreflightConfig) Initialize() error {
	if p.IsExistOnDisk() {
		return nil
	}
	p.AddMinK8sV("1.15")
	p.AddMinMongoV("3.6")
	p.AddImage("nginx", "nginx")
	p.AddImage("netcat", "subfuzion/netcat")
	p.AddImage("mongo", "mongo")
	return p.Write()
}
