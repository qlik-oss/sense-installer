package qliksense

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/cli/opts"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"

	"strings"

	"golang.org/x/net/context"
	yaml "gopkg.in/yaml.v2"
)

// Images ...
type Images struct {
	Images []string `yaml:"images"`
}

// PullImages ...
func (p *Qliksense) PullImages() error {
	var (
		image       string
		err         error
		yamlVersion string
		valid       bool
		images      Images
	)

	if yamlVersion, err = p.CallPorter([]string{"invoke", "--action", "about"},
		func(x string) (out *string) {
			if strings.HasPrefix(x, "qlikSenseVersion") {
				valid = true
			}
			if strings.HasPrefix(x, "execution") {
				valid = false
			}
			if valid {
				return &x
			}
			return nil
		}); err != nil {
		return err
	}

	if err = yaml.Unmarshal([]byte(yamlVersion), &images); err != nil {
		return err
	}
	for _, image = range images.Images {
		if err = p.PullImage(image); err != nil {
			fmt.Print(err)
		}
		println("---")
	}

	return nil
}

// PullImage ...
func (p *Qliksense) PullImage(imageName string) error {
	var (
		cli          *command.DockerCli
		dockerOutput io.Writer
		response     io.ReadCloser
		pullOptions  types.ImagePullOptions
		ctx          context.Context
		ref          reference.Named
		repoInfo     *registry.RepositoryInfo
		authConfig   types.AuthConfig
		encodedAuth  string
		termFd       uintptr
		err          error
	)
	// TODO: Create a real cli config context
	ctx = context.Background()
	if cli, err = command.NewDockerCli(); err != nil {
		return err
	}

	if err = cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return err
	}

	if ref, err = reference.ParseNormalizedNamed(imageName); err != nil {
		return err
	}
	if repoInfo, err = registry.ParseRepositoryInfo(ref); err != nil {
		return err
	}

	authConfig = command.ResolveAuthConfig(ctx, cli, repoInfo.Index)
	if encodedAuth, err = command.EncodeAuthToBase64(authConfig); err != nil {
		return err
	}

	pullOptions = types.ImagePullOptions{
		RegistryAuth: encodedAuth,
	}

	if response, err = cli.Client().ImagePull(ctx, imageName, pullOptions); err != nil {
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

// PullImage ...
func (p *Qliksense) TagAndPushImages(registry string) error {
	var (
		image       string
		err         error
		yamlVersion string
		valid       bool
		images      Images
	)

	if yamlVersion, err = p.CallPorter([]string{"invoke", "--action", "about"},
		func(x string) (out *string) {
			if strings.HasPrefix(x, "qlikSenseVersion") {
				valid = true
			}
			if strings.HasPrefix(x, "execution") {
				valid = false
			}
			if valid {
				return &x
			}
			return nil
		}); err != nil {
		return err
	}

	if err = yaml.Unmarshal([]byte(yamlVersion), &images); err != nil {
		return err
	}

	for _, image = range images.Images {
		if err = p.TagAndPush(image, registry); err != nil {
			fmt.Print(err)
		}
		println("---")
	}

	return nil
}

// PullImage ...
func (p *Qliksense) TagAndPush(image string, registryName string) error {
	var (
		cli              *command.DockerCli
		dockerOutput     io.Writer
		response         io.ReadCloser
		pushOptions      types.ImagePushOptions
		ctx              context.Context
		newName          string
		segments         []string
		imageList        []types.ImageSummary
		imageListOptions types.ImageListOptions
		filter           opts.FilterOpt
		filters          filters.Args
		ref              reference.Named
		repoInfo         *registry.RepositoryInfo
		authConfig       types.AuthConfig
		encodedAuth      string
		termFd           uintptr
		err              error
	)
	// TODO: Create a real cli config context
	ctx = context.Background()
	if cli, err = command.NewDockerCli(); err != nil {
		return err
	}
	if err = cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return err
	}
	segments = strings.Split(image, "/")
	newName = registryName + "/" + segments[len(segments)-1]

	filters = filter.Value()
	filters.Add("reference", image)
	imageListOptions = types.ImageListOptions{
		Filters: filters,
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
