package hellokubernetes

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	helloKubernetesLong = templates.LongDesc(i18n.T(`
	Print resource information.

	This utility demonstrates how to build custom Kubernetes commands 
	while maintaining Kubernetes CLI conventions, such as working with 
	resources, handling input/output streams, and adhering to the 
	kubectl subcommand structure.`))

	hellokubernetesExample = templates.Examples(`
		# Print "Hello <kind of resource> <name of resource>
		kubectl hello-kubernetes -f file

		Print "Hello <kind of resource> <name of resource> <creation time>
		kubectl hello-kubernetes type/name
	`)
)

type HelloKubernetesOptions struct {
	// Filename options
	resource.FilenameOptions
	genericiooptions.IOStreams

	// TODO - add print flags
	// PrintFlags *genericclioptions.PrintFlags
}

func NewHelloKubernetesOptions(ioStreams genericiooptions.IOStreams) *HelloKubernetesOptions {
	return &HelloKubernetesOptions{
		IOStreams: ioStreams,
	}
}

func NewCmdHelloKubernetes(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := NewHelloKubernetesOptions(ioStreams)

	cmd := &cobra.Command{
		Use:     "hello-kubernetes (-f FILENAME | TYPE NAME)",
		Short:   i18n.T("Print resource information"),
		Long:    helloKubernetesLong,
		Example: hellokubernetesExample,
		Run: func(cmd *cobra.Command, args []string) {
			// cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(args))
			// cmdutil.CheckErr(o.RunHelloKubernetes())
		},
	}

	usage := "identifying the resource to retrieve and print its information"
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)

	return cmd
}

// Complete adapts from the command line args and factory to the data required.
// TODO - Add --output flag later and use this funciton to bind the --output
func (o HelloKubernetesOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	return nil
}

func (o HelloKubernetesOptions) Validate(args []string) error {
	// Ensure either filename is passed or resource type/name
	if len(args) > 0 && !cmdutil.IsFilenameSliceEmpty(o.FilenameOptions.Filenames, o.FilenameOptions.Kustomize) {
		return fmt.Errorf("cannot provide both arguments (type/name) and file input")
	} else if len(args) == 0 && cmdutil.IsFilenameSliceEmpty(o.FilenameOptions.Filenames, o.FilenameOptions.Kustomize) {
		return fmt.Errorf("must provide either arguments (type/name) and file input")
	}
	return nil
}

func (o HelloKubernetesOptions) RunHelloKubernetes() error {
	fmt.Fprintln(o.Out, "Hello World")
	return nil
}
