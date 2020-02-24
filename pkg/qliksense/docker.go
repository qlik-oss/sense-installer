package qliksense

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	imageTypes "github.com/containers/image/v5/types"
	qapi "github.com/qlik-oss/sense-installer/pkg/api"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
)

type imageNameParts struct {
	name string
	tag  string
}

const (
	imagesDirName           = "images"
	imageIndexDirName       = "index"
	imageSharedBlobsDirName = "blobs"
)

// PullImages ...
func (q *Qliksense) PullImagesForCurrentCR() error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		return err
	}
	version := qcr.GetLabelFromCr("version")
	profile := qcr.Spec.Profile
	repoDir := qcr.Spec.ManifestsRoot

	imagesDir, err := q.setupImagesDir()
	if err != nil {
		return err
	}

	versionOut, stored, err := q.readOrGenerateVersionOutput(imagesDir, version, repoDir, profile)
	if err != nil {
		return err
	}

	for _, image := range versionOut.Images {
		if err := q.pullImage(image, imagesDir); err != nil {
			fmt.Printf("%v\n", err)
			return err
		}
		fmt.Print("---\n")
	}

	if version != "" && !stored {
		if err := q.writeVersionOutput(versionOut, imagesDir, version); err != nil {
			return err
		}
	}
	return nil
}

func (q *Qliksense) pullImage(image, imagesDir string) error {
	srcRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%v", image))
	if err != nil {
		return err
	}
	nameTag := q.getImageNameParts(image)
	targetDir := filepath.Join(imagesDir, imageIndexDirName, nameTag.name, nameTag.tag)
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}

	fmt.Printf("==> Pulling image %v:%v\n", nameTag.name, nameTag.tag)
	destRef, err := alltransports.ParseImageName(fmt.Sprintf("oci:%v", targetDir))
	if err != nil {
		return err
	}

	policyContext, err := signature.NewPolicyContext(&signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}})
	if err != nil {
		return err
	}
	defer policyContext.Destroy()

	if _, err := copy.Image(context.Background(), policyContext, destRef, srcRef, &copy.Options{
		ReportWriter: os.Stdout,
		SourceCtx: &imageTypes.SystemContext{
			ArchitectureChoice: "amd64",
			OSChoice:           "linux",
		},
		DestinationCtx: &imageTypes.SystemContext{
			OCISharedBlobDirPath: filepath.Join(imagesDir, imageSharedBlobsDirName),
		},
	}); err != nil {
		return err
	}
	return nil
}

// TagAndPushImages ...
func (q *Qliksense) PushImagesForCurrentCR(registry string) error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		return err
	}
	version := qcr.GetLabelFromCr("version")
	profile := qcr.Spec.Profile
	repoDir := qcr.Spec.ManifestsRoot

	imagesDir, err := q.setupImagesDir()
	if err != nil {
		return err
	}

	versionOut, stored, err := q.readOrGenerateVersionOutput(imagesDir, version, repoDir, profile)
	if err != nil {
		return err
	}

	for _, image := range versionOut.Images {
		if err = q.pushImage(image, imagesDir, registry); err != nil {
			fmt.Printf("%v\n", err)
			return err
		}
		fmt.Print("---\n")
	}

	if version != "" && !stored {
		if err := q.writeVersionOutput(versionOut, imagesDir, version); err != nil {
			return err
		}
	}

	return nil
}

func (q *Qliksense) pushImage(image, imagesDir, registryName string) error {
	imageNameParts := q.getImageNameParts(image)
	srcDir := filepath.Join(imagesDir, imageIndexDirName, imageNameParts.name, imageNameParts.tag)
	if exists, err := q.directoryExists(srcDir); err != nil {
		return err
	} else if !exists {
		if err := q.pullImage(image, imagesDir); err != nil {
			return err
		}
	}
	srcRef, err := alltransports.ParseImageName(fmt.Sprintf("oci:%v", srcDir))
	if err != nil {
		return err
	}

	newImage := fmt.Sprintf("%v/%v:%v", registryName, imageNameParts.name, imageNameParts.tag)
	fmt.Printf("==> Pushing image: %v\n", newImage)

	destRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%v", newImage))
	if err != nil {
		return err
	}

	policyContext, err := signature.NewPolicyContext(&signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}})
	if err != nil {
		return err
	}
	defer policyContext.Destroy()

	if _, err = copy.Image(context.Background(), policyContext, destRef, srcRef, &copy.Options{
		ReportWriter: os.Stdout,
		SourceCtx: &imageTypes.SystemContext{
			OCISharedBlobDirPath: filepath.Join(imagesDir, imageSharedBlobsDirName),
		},
		DestinationCtx: &imageTypes.SystemContext{
			DockerInsecureSkipTLSVerify: imageTypes.OptionalBoolTrue,
		},
	}); err != nil {
		return err
	}
	return nil
}

func (q *Qliksense) directoryExists(path string) (exists bool, err error) {
	if info, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		exists = false
		err = nil
	} else if err != nil && !os.IsNotExist(err) {
		exists = false
	} else if err == nil && info.IsDir() {
		exists = true
	} else if err == nil && !info.IsDir() {
		exists = false
		err = fmt.Errorf("path: %v is occupied by a file instead of a directory", path)
	}
	return exists, err
}

func (q *Qliksense) getImageNameParts(image string) imageNameParts {
	segments := strings.Split(image, "/")
	nameTag := strings.Split(segments[len(segments)-1], ":")
	if len(nameTag) < 2 {
		nameTag = append(nameTag, "latest")
	}
	return imageNameParts{
		name: nameTag[0],
		tag:  nameTag[1],
	}
}

func (q *Qliksense) setupImagesDir() (string, error) {
	imagesDir := filepath.Join(q.QliksenseHome, imagesDirName)

	imageIndexDir := filepath.Join(imagesDir, imageIndexDirName)
	if err := os.MkdirAll(imageIndexDir, os.ModePerm); err != nil {
		return "", err
	}

	sharedBlobsDir := filepath.Join(imagesDir, imageSharedBlobsDirName)
	if err := os.MkdirAll(sharedBlobsDir, os.ModePerm); err != nil {
		return "", err
	}

	return imagesDir, nil
}

func (q *Qliksense) readOrGenerateVersionOutput(imagesDir, version, repoDir, profile string) (versionOut *VersionOutput, stored bool, err error) {
	if version != "" {
		versionOut, err = q.readVersionOutput(imagesDir, version)
		if versionOut != nil {
			stored = true
		}
	}
	if versionOut == nil {
		if versionOut, err = q.AboutDir(repoDir, profile); err != nil {
			return nil, false, err
		}
	}
	return versionOut, stored, nil
}

func (q *Qliksense) readVersionOutput(imagesDir, version string) (*VersionOutput, error) {
	var versionOut VersionOutput
	versionFile := filepath.Join(imagesDir, version)
	if versionOutBytes, err := ioutil.ReadFile(versionFile); err != nil {
		return nil, err
	} else if err = yaml.Unmarshal(versionOutBytes, &versionOut); err != nil {
		return nil, err
	}
	return &versionOut, nil
}

func (q *Qliksense) writeVersionOutput(versionOut *VersionOutput, imagesDir, version string) error {
	versionFile := filepath.Join(imagesDir, version)
	if versionOutBytes, err := yaml.Marshal(versionOut); err != nil {
		return err
	} else if err = ioutil.WriteFile(versionFile, versionOutBytes, os.ModePerm); err != nil {
		return err
	}
	return nil
}
