package qliksense

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	imageTypes "github.com/containers/image/v5/types"
	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"golang.org/x/net/context"
	yaml "gopkg.in/yaml.v2"
)

// PullImages ...
func (p *Qliksense) PullImages(gitRef, profile string, engine bool) error {
	var (
		image, versionFile, imagesDir, homeDir string
		err                                    error
		versionOut                             *VersionOutput
	)
	println("getting images list...")

	// TODO: get getref and profile from config/cr for About function call
	if versionOut, err = p.About(gitRef, profile); err != nil {
		return err
	}

	if homeDir, err = homedir.Dir(); err != nil {
		return err
	}
	imagesDir = filepath.Join(homeDir, ".qliksense", "images")
	os.MkdirAll(imagesDir, 0644)
	versionFile = filepath.Join(imagesDir, versionOut.QliksenseVersion)

	if _, err = os.Stat(versionFile); err != nil {
		if os.IsNotExist(err) {
			if yamlVersion, err := yaml.Marshal(versionOut); err != nil {
				return err
			} else if err = ioutil.WriteFile(versionFile, yamlVersion, 0644); err != nil {
				return err
			}
		} else {
			return errors.Errorf("Unable to determine About file %v exists", versionFile)
		}
	}
	for _, image = range versionOut.Images {
		if _, err = p.PullImage(image, engine); err != nil {
			fmt.Print(err)
		}
		println("---")
	}

	return nil
}

// PullImage ...
func (p *Qliksense) PullImage(imageName string, engine bool) (map[string]string, error) {
	if engine {
		return p.pullDockerImage(imageName)
	}
	return p.pullImage(imageName)
}

func (p *Qliksense) commandTimeoutContext(commandTimeout time.Duration) (context.Context, context.CancelFunc) {
	ctx := context.Background()
	var cancel context.CancelFunc = func() {}
	if commandTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, commandTimeout)
	}
	return ctx, cancel
}

func (p *Qliksense) pullImage(imageName string) (map[string]string, error) {
	var (
		ctx                         context.Context
		cancel                      context.CancelFunc
		srcRef, destRef             imageTypes.ImageReference
		blobDir, targetDir, homeDir string
		segments                    []string
		nameTag                     []string
		err                         error
		policyContext               *signature.PolicyContext
	)
	ctx, cancel = p.commandTimeoutContext(0)
	defer cancel()

	if srcRef, err = alltransports.ParseImageName("docker://" + imageName); err != nil {
		return nil, err
	}
	segments = strings.Split(imageName, "/")
	nameTag = strings.Split(segments[len(segments)-1], ":")
	if len(nameTag) < 2 {
		nameTag = append(nameTag, "latest")
	}
	if homeDir, err = homedir.Dir(); err != nil {
		return nil, err
	}
	targetDir = filepath.Join(homeDir, ".qliksense", "images", nameTag[0], nameTag[1])

	fmt.Printf("==> Pulling image %v:%v", nameTag[0], nameTag[1])
	fmt.Println()
	os.MkdirAll(targetDir, 0644)
	blobDir = filepath.Join(homeDir, ".qliksense", "blobs")
	os.MkdirAll(blobDir, 0644)

	if destRef, err = alltransports.ParseImageName("oci:" + targetDir); err != nil {
		return nil, err
	}

	if policyContext, err = signature.NewPolicyContext(&signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}); err != nil {
		return nil, err
	}
	defer policyContext.Destroy()

	_, err = copy.Image(ctx, policyContext, destRef, srcRef, &copy.Options{
		ReportWriter: os.Stdout,
		SourceCtx: &imageTypes.SystemContext{
			ArchitectureChoice: "amd64",
			OSChoice:           "linux",
		},
		DestinationCtx: &imageTypes.SystemContext{
			OCISharedBlobDirPath: blobDir,
		},
	})
	return nil, err
}
func (p *Qliksense) pullDockerImage(imageName string) (map[string]string, error) {
	var (
		cli          *command.DockerCli
		dockerOutput io.Writer
		response     io.ReadCloser
		pullOptions  types.ImagePullOptions
		ctx          context.Context
		cancel       context.CancelFunc
		ref          reference.Named
		repoInfo     *registry.RepositoryInfo
		authConfig   types.AuthConfig
		encodedAuth  string
		termFd       uintptr
		err          error
	)
	ctx, cancel = p.commandTimeoutContext(0)
	defer cancel()

	if cli, err = command.NewDockerCli(); err != nil {
		return nil, err
	}

	if err = cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return nil, err
	}

	if ref, err = reference.ParseNormalizedNamed(imageName); err != nil {
		return nil, err
	}
	if repoInfo, err = registry.ParseRepositoryInfo(ref); err != nil {
		return nil, err
	}

	authConfig = command.ResolveAuthConfig(ctx, cli, repoInfo.Index)
	if encodedAuth, err = command.EncodeAuthToBase64(authConfig); err != nil {
		return nil, err
	}

	pullOptions = types.ImagePullOptions{
		RegistryAuth: encodedAuth,
	}

	if response, err = cli.Client().ImagePull(ctx, imageName, pullOptions); err != nil {
		return nil, err
	}
	defer response.Close()

	dockerOutput = ioutil.Discard
	// if b.IsVerbose() {
	// 	dockerOutput = b.Out
	// }
	dockerOutput = os.Stdout
	termFd, _ = term.GetFdInfo(dockerOutput)
	// Setting this to false here because Moby os.Exit(1) all over the place and this fails on WSL (only)
	// when Term is true.
	isTerm := false
	if err = jsonmessage.DisplayJSONMessagesStream(response, dockerOutput, termFd, isTerm, nil); err != nil {
		return nil, err
	}
	inspectData, _, err := cli.Client().ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return nil, err
	}
	return inspectData.ContainerConfig.Labels, nil
}

//TagAndPushImages ...
func (p *Qliksense) TagAndPushImages(registry string, engine bool) error {
	var (
		image       string
		err         error
		yamlVersion string
		images      VersionOutput
	)

	if err = yaml.Unmarshal([]byte(yamlVersion), &images); err != nil {
		return err
	}

	for _, image = range images.Images {
		if err = p.TagAndPush(image, registry, engine); err != nil {
			fmt.Print(err)
		}
		println("---")
	}

	return nil
}

func (p *Qliksense) directoryExists(path string) (exists bool, err error) {
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

//TagAndPush ...
func (p *Qliksense) TagAndPush(image string, registryName string, engine bool) error {
	if engine {
		return p.tagAndDockerPush(image, registryName)
	}
	return p.tagAndPush(image, registryName)
}

func (p *Qliksense) tagAndPush(image string, registryName string) error {
	var (
		ctx                               context.Context
		cancel                            context.CancelFunc
		srcRef, destRef                   imageTypes.ImageReference
		blobDir, srcDir, homeDir, newName string
		segments                          []string
		nameTag                           []string
		err                               error
		policyContext                     *signature.PolicyContext
		srcExists                         bool
	)
	ctx, cancel = p.commandTimeoutContext(0)
	defer cancel()

	segments = strings.Split(image, "/")
	nameTag = strings.Split(segments[len(segments)-1], ":")

	if len(nameTag) < 2 {
		nameTag = append(nameTag, "latest")
	}
	if homeDir, err = homedir.Dir(); err != nil {
		return err
	}
	srcDir = filepath.Join(homeDir, ".qliksense", "images", nameTag[0], nameTag[1])
	if srcExists, err = p.directoryExists(srcDir); err != nil {
		return err
	}
	if !srcExists {
		if _, err = p.PullImage(image, false); err != nil {
			return err
		}
	}
	if srcRef, err = alltransports.ParseImageName("oci:" + srcDir); err != nil {
		return err
	}

	if segments[0] == "docker.io" {
		image = strings.Join(segments[1:], "/")
	}
	newName = "//" + registryName + "/" + segments[len(segments)-1]

	fmt.Printf("==> Tag and push image to %v", newName)
	fmt.Println()

	if destRef, err = alltransports.ParseImageName("docker:" + newName); err != nil {
		return err
	}

	if policyContext, err = signature.NewPolicyContext(&signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}); err != nil {
		return err
	}
	defer policyContext.Destroy()

	blobDir = filepath.Join(homeDir, ".qliksense", "blobs")
	os.MkdirAll(blobDir, 0644)

	_, err = copy.Image(ctx, policyContext, destRef, srcRef, &copy.Options{
		ReportWriter: os.Stdout,
		SourceCtx: &imageTypes.SystemContext{
			OCISharedBlobDirPath: blobDir,
		},
		DestinationCtx: &imageTypes.SystemContext{
			DockerDaemonInsecureSkipTLSVerify: true,
		},
	})
	return err
}

// PullImage ...
func (p *Qliksense) tagAndDockerPush(image string, registryName string) error {
	var (
		cli              *command.DockerCli
		dockerOutput     io.Writer
		response         io.ReadCloser
		pushOptions      types.ImagePushOptions
		ctx              context.Context
		cancel           context.CancelFunc
		newName          string
		segments         []string
		imageList        []types.ImageSummary
		imageListOptions types.ImageListOptions
		filterArgs       filters.Args
		ref              reference.Named
		repoInfo         *registry.RepositoryInfo
		authConfig       types.AuthConfig
		encodedAuth      string
		termFd           uintptr
		err              error
	)
	// TODO: Create a real cli config context
	ctx, cancel = p.commandTimeoutContext(0)
	defer cancel()
	if cli, err = command.NewDockerCli(); err != nil {
		return err
	}
	if err = cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return err
	}
	segments = strings.Split(image, "/")
	if segments[0] == "docker.io" {
		image = strings.Join(segments[1:], "/")
	}
	newName = registryName + "/" + segments[len(segments)-1]

	filterArgs = filters.NewArgs()
	filterArgs.Add("reference", image)
	imageListOptions = types.ImageListOptions{
		Filters: filterArgs,
	}
	if imageList, err = cli.Client().ImageList(ctx, imageListOptions); err != nil {
		return err
	}
	if imageList == nil || len(imageList) <= 0 {
		fmt.Printf("Use `qliksense pull`, to pull %v for an air gap push", newName)
		return nil
	}

	if err = cli.Client().ImageTag(ctx, image, newName); err != nil {
		return err
	}

	if ref, err = reference.ParseNormalizedNamed(image); err != nil {
		return err
	}
	if repoInfo, err = registry.ParseRepositoryInfo(ref); err != nil {
		return err
	}
	authConfig = command.ResolveAuthConfig(ctx, cli, repoInfo.Index)
	if encodedAuth, err = command.EncodeAuthToBase64(authConfig); err != nil {
		return err
	}
	pushOptions = types.ImagePushOptions{
		All:          true,
		RegistryAuth: encodedAuth,
	}

	if response, err = cli.Client().ImagePush(ctx, newName, pushOptions); err != nil {
		return err
	}
	defer response.Close()

	dockerOutput = ioutil.Discard
	// if b.IsVerbose() {
	// 	dockerOutput = b.Out
	// }
	dockerOutput = os.Stdout
	termFd, _ = term.GetFdInfo(dockerOutput)
	// Setting this to false here because Moby os.Exit(1) all over the place and this fails on WSL (only)
	// when Term is true.
	isTerm := false
	if err = jsonmessage.DisplayJSONMessagesStream(response, dockerOutput, termFd, isTerm, nil); err != nil {
		return err
	}
	return nil
}
