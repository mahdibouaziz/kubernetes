package helloworld

import (
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericiooptions"

	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	helloworlExample = templates.Examples(`
		# Print Hello World
		kubectl hello-world
	`)
)

type HelloWorldOptions struct {
	genericiooptions.IOStreams
}

func NewHelloWorldOptions(ioStreams genericiooptions.IOStreams) *HelloWorldOptions {
	return &HelloWorldOptions{
		IOStreams: ioStreams,
	}
}

func NewCmdHelloWorld(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := NewHelloWorldOptions(ioStreams)

	cmd := &cobra.Command{
		Use:     "hello-world",
		Short:   i18n.T("Print hello world"),
		Long:    i18n.T("Print hello world."),
		Example: helloworlExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.RunHelloWorld())
		},
	}

	return cmd
}

func (o HelloWorldOptions) RunHelloWorld() error {
	fmt.Fprintln(o.Out, "Hello World")
	return nil
}
