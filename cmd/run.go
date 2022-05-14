/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/joyme123/kubectl-tools/remote"
	"github.com/joyme123/kubectl-tools/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

var opts *RunOptions

// RunOptions ...
type RunOptions struct {
	ConfigFlags *genericclioptions.ConfigFlags
	Pod         string
	Container   string
}

// ToKubeRequest converts RunOptions to KubeRequest
func (o *RunOptions) ToKubeRequest() (*remote.KubeRequest, error) {
	restConfig, err := o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	clientset := kubernetes.NewForConfigOrDie(restConfig)

	return &remote.KubeRequest{
		Clientset:  clientset,
		RestConfig: restConfig,
		Namespace:  *o.ConfigFlags.Namespace,
		Pod:        o.Pod,
		Container:  o.Container,
	}, nil
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Println("you should specify a tool name")
			os.Exit(0)
		}

		if t, ok := tools.Set[args[0]]; !ok {
			log.Warning("tool %s doesn't support now, you can create an issue or pull request on github.com/joyme123/kubectl-tools\n", args[0])
		} else {
			kube, err := opts.ToKubeRequest()
			if err != nil {
				panic(err)
			}
			err = remote.Run(kube, t, args)
			if err != nil {
				panic(err)
			}
		}
	},
}

func init() {
	opts = &RunOptions{
		ConfigFlags: genericclioptions.NewConfigFlags(true),
	}
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&opts.Pod, "pod", "p", "", "pod name")
	runCmd.Flags().StringVarP(&opts.Container, "container", "c", "", "container name")
	opts.ConfigFlags.AddFlags(runCmd.Flags())

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
