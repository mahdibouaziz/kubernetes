package hellokubernetes

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
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
	// arguments
	args []string

	// Configure the builder
	builder *resource.Builder

	// configure the printer
	PrintFlags *genericclioptions.PrintFlags
	ToPrinter  func(string) (printers.ResourcePrinter, error)

	// generic IOStreams
	genericiooptions.IOStreams
}

func NewHelloKubernetesOptions(ioStreams genericiooptions.IOStreams) *HelloKubernetesOptions {
	return &HelloKubernetesOptions{
		IOStreams:  ioStreams,
		PrintFlags: genericclioptions.NewPrintFlags("greeted").WithTypeSetter(scheme.Scheme),
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
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.RunHelloKubernetes())
		},
	}

	// Add the required flags
	o.AddFlags(cmd, ioStreams)

	return cmd
}

// AddFlags Adds the required Flags for the command
func (o *HelloKubernetesOptions) AddFlags(cmd *cobra.Command, ioStreams genericiooptions.IOStreams) {
	// printer flags
	o.PrintFlags.AddFlags(cmd)

	// Filename flags
	usage := "identifying the resource to retrieve and print its information"
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)
}

// Complete adapts from the command line args and factory to the data required.
func (o *HelloKubernetesOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	o.args = args
	o.builder = f.NewBuilder()

	// Configure the printer
	o.ToPrinter = func(operation string) (printers.ResourcePrinter, error) {
		o.PrintFlags.NamePrintFlags.Operation = operation
		cmdutil.PrintFlagsWithDryRunStrategy(o.PrintFlags, cmdutil.DryRunNone)
		return o.PrintFlags.ToPrinter()
	}
	return nil
}

// Validate checks to the HelloKubernetesOptions to see if there is sufficient information run the command.
func (o *HelloKubernetesOptions) Validate() error {
	// Ensure either filename is passed or resource type/name
	if len(o.args) > 0 && !cmdutil.IsFilenameSliceEmpty(o.FilenameOptions.Filenames, o.FilenameOptions.Kustomize) {
		return fmt.Errorf("cannot provide both arguments (type/name) and file input (--filename or -f)")
	} else if len(o.args) == 0 && cmdutil.IsFilenameSliceEmpty(o.FilenameOptions.Filenames, o.FilenameOptions.Kustomize) {
		return fmt.Errorf("must provide either arguments (type/name) or file input (--filename or -f)")
	}
	return nil
}

// RunHelloKubernetes Does the work
func (o *HelloKubernetesOptions) RunHelloKubernetes() error {

	if !cmdutil.IsFilenameSliceEmpty(o.FilenameOptions.Filenames, o.FilenameOptions.Kustomize) {
		// handle filename was passed
		r := o.builder.
			Unstructured().
			ContinueOnError().
			DefaultNamespace().
			FilenameParam(false, &o.FilenameOptions).
			Flatten().
			Do()
		err := r.Err()
		if err != nil {
			return err
		}

		err = r.Visit(func(info *resource.Info, err error) error {
			if err != nil {
				return err
			}
			resultMsg := fmt.Sprintf("Hello %s %s \n", info.Mapping.GroupVersionKind.Kind, info.Name)
			printer, err := o.ToPrinter(resultMsg)
			if err != nil {
				return err
			}
			return printer.PrintObj(info.Object, o.Out)

		})
		if err != nil {
			return err
		}

	} else {
		// handle resource type/name was passed
		b := o.builder.
			Unstructured().
			ContinueOnError().
			NamespaceParam("default").
			// NamespaceParam(o.namespace).
			DefaultNamespace().
			Flatten().
			ResourceTypeOrNameArgs(false, o.args...).
			Latest()

		r := b.Do()

		return r.Visit(func(info *resource.Info, err error) error {
			if err != nil {
				return err
			}

			// Get metadata for the object
			metaObj, err := meta.Accessor(info.Object)
			if err != nil {
				return fmt.Errorf("unable to access metadata: %v", err)
			}

			resultMsg := fmt.Sprintf("Hello %s %s %s\n", info.Mapping.GroupVersionKind.Kind, info.Name, metaObj.GetCreationTimestamp().Time)
			printer, err := o.ToPrinter(resultMsg)
			if err != nil {
				return err
			}
			return printer.PrintObj(info.Object, o.Out)
		})
	}

	return nil
}
