package qliksense

import (
	"errors"
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

func (q *Qliksense) PullImages(version, profile string) error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	if version != "" {
		if !qConfig.IsRepoExistForCurrent(version) {
			if err := q.FetchQK8s(version); err != nil {
				return err
			}
		}
	}
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		return err
	}
	if !qcr.IsRepoExist() {
		return errors.New("ManifestsRoot not found")
	}
	if profile != "" {
		qcr.Spec.Profile = profile
		if err := qConfig.WriteCR(qcr); err != nil {
			return err
		}
	}
	return q.PullImagesForCurrentCR()
}

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

	imagesDir, err := setupImagesDir(q.QliksenseHome)
	if err != nil {
		return err
	}

	versionOut, stored, err := q.readOrGenerateVersionOutput(imagesDir, version, repoDir, profile)
	if err != nil {
		return err
	}

	images := versionOut.Images
	if err := q.appendAdditionalImages(&images, qcr); err != nil {
		return err
	}

	for _, image := range images {
		if err := pullImage(image, imagesDir); err != nil {
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

func (q *Qliksense) appendOpsRunnerImage(images *[]string, qcr *qapi.QliksenseCR) {
	if qcr.Spec.OpsRunner != nil && qcr.Spec.OpsRunner.Image != "" {
		*images = append(*images, qcr.Spec.OpsRunner.Image)
	}
}

func (q *Qliksense) appendPreflightImages(images *[]string) {
	pf := qapi.NewPreflightConfig(q.QliksenseHome)
	for _, preflightImage := range pf.GetImageMap() {
		*images = append(*images, preflightImage)
	}
}

func (q *Qliksense) appendOperatorImages(images *[]string) error {
	if operatorImages, err := getImageList([]byte(q.GetOperatorControllerString())); err != nil {
		return err
	} else {
		*images = append(*images, operatorImages...)
		return nil
	}
}

func pullImage(image, imagesDir string) error {
	srcRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%v", image))
	if err != nil {
		return err
	}
	nameTag := getImageNameParts(image)
	targetDir := filepath.Join(imagesDir, imageIndexDirName, nameTag.name, nameTag.tag)
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}

	destRef, err := alltransports.ParseImageName(fmt.Sprintf("oci:%v", targetDir))
	if err != nil {
		return err
	}

	policyContext, err := signature.NewPolicyContext(&signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}})
	if err != nil {
		return err
	}
	defer policyContext.Destroy()

	fmt.Printf("==> Pulling image from %v\n", srcRef.StringWithinTransport())
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

func (q *Qliksense) PushImagesForCurrentCR() error {
	qConfig := qapi.NewQConfig(q.QliksenseHome)
	qcr, err := qConfig.GetCurrentCR()
	if err != nil {
		return err
	}
	version := qcr.GetLabelFromCr("version")
	profile := qcr.Spec.Profile
	repoDir := qcr.Spec.ManifestsRoot

	dockerConfigJsonSecret, err := qConfig.GetPushDockerConfigJsonSecret()
	if err != nil {
		if os.IsNotExist(err) {
			dockerConfigJsonSecret = &qapi.DockerConfigJsonSecret{
				Uri: qcr.Spec.GetImageRegistry(),
			}
		} else {
			return err
		}
	}

	imagesDir, err := setupImagesDir(q.QliksenseHome)
	if err != nil {
		return err
	}

	versionOut, stored, err := q.readOrGenerateVersionOutput(imagesDir, version, repoDir, profile)
	if err != nil {
		return err
	}

	images := versionOut.Images
	if err := q.appendAdditionalImages(&images, qcr); err != nil {
		return err
	}

	for _, image := range images {
		if err = pushImage(image, imagesDir, dockerConfigJsonSecret); err != nil {
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

func (q *Qliksense) appendAdditionalImages(images *[]string, qcr *qapi.QliksenseCR) error {
	if err := q.appendOperatorImages(images); err != nil {
		return err
	}
	q.appendOpsRunnerImage(images, qcr)
	q.appendPreflightImages(images)
	return nil
}

func pushImage(image, imagesDir string, dockerConfigJsonSecret *qapi.DockerConfigJsonSecret) error {
	imageNameParts := getImageNameParts(image)
	srcDir := filepath.Join(imagesDir, imageIndexDirName, imageNameParts.name, imageNameParts.tag)
	if exists, err := directoryExists(srcDir); err != nil {
		return err
	} else if !exists {
		if err := pullImage(image, imagesDir); err != nil {
			return err
		}
	}
	srcRef, err := alltransports.ParseImageName(fmt.Sprintf("oci:%v", srcDir))
	if err != nil {
		return err
	}

	newImage := fmt.Sprintf("%v/%v:%v", dockerConfigJsonSecret.Uri, imageNameParts.name, imageNameParts.tag)
	destRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%v", newImage))
	if err != nil {
		return err
	}

	policyContext, err := signature.NewPolicyContext(&signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}})
	if err != nil {
		return err
	}
	defer policyContext.Destroy()

	destinationCtx := &imageTypes.SystemContext{
		DockerInsecureSkipTLSVerify: imageTypes.OptionalBoolTrue,
	}
	if dockerConfigJsonSecret.Username != "" {
		destinationCtx.DockerAuthConfig = &imageTypes.DockerAuthConfig{
			Username: dockerConfigJsonSecret.Username,
			Password: dockerConfigJsonSecret.Password,
		}
	}
	fmt.Printf("==> Pushing image to: %v\n", destRef.StringWithinTransport())
	if _, err = copy.Image(context.Background(), policyContext, destRef, srcRef, &copy.Options{
		ReportWriter: os.Stdout,
		SourceCtx: &imageTypes.SystemContext{
			OCISharedBlobDirPath: filepath.Join(imagesDir, imageSharedBlobsDirName),
		},
		DestinationCtx: destinationCtx,
	}); err != nil {
		return err
	}
	return nil
}

func directoryExists(path string) (exists bool, err error) {
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

func getImageNameParts(image string) imageNameParts {
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

func setupImagesDir(qliksenseHome string) (string, error) {
	imagesDir := filepath.Join(qliksenseHome, imagesDirName)

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
