package main

import (
	// "fmt"

	// "github.com/deislabs/porter/pkg/mixin"
	// "github.com/deislabs/porter/pkg/porter"
	// "github.com/spf13/cobra"
)

// func installPorterMixin(p *porter.Porter) *cobra.Command {
// 	opts := mixin.InstallOptions{}
// 	cmd := &cobra.Command{
// 		Use:   "install NAME",
// 		Short: "Install a mixin",
// 		Example: `  qliksense install helm --url https://cdn.deislabs.io/porter/mixins/helm
//   qliksense install helm --feed-url https://cdn.deislabs.io/porter/atom.xml
//   qliksense install azure --version v0.4.0-ralpha.1+dubonnet --url https://cdn.deislabs.io/porter/mixins/azure
//   qliksense install kubernetes --version canary --url https://cdn.deislabs.io/porter/mixins/kubernetes`,
// 		PreRunE: func(cmd *cobra.Command, args []string) error {
// 			return opts.Validate(args)
// 		},
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return p.InstallMixin(opts)

// 		},
// 	}

// 	cmd.Flags().StringVarP(&opts.Version, "version", "v", "latest",
// 		"The mixin version. This can either be a version number, or a tagged release like 'latest' or 'canary'")
// 	cmd.Flags().StringVar(&opts.URL, "url", "",
// 		"URL from where the mixin can be downloaded, for example https://github.com/org/proj/releases/downloads")
// 	cmd.Flags().StringVar(&opts.FeedURL, "feed-url", "",
// 		fmt.Sprintf(`URL of an atom feed where the mixin can be downloaded (default %s)`, mixin.DefaultFeedUrl))
// 	return cmd
// }
