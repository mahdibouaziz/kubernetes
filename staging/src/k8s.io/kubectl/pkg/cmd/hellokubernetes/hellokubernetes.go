package hellokubernetes

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

	// namespace options
	namespace        string
	enforceNamespace bool
	allNamespaces    bool

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
	cmd.Flags().BoolVarP(&o.allNamespaces, "all-namespaces", "A", o.allNamespaces, "If true, check the specified action in all namespaces.")

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

	// Configure the namespace
	var err error
	o.namespace, o.enforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
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

	// Create the builder
	b := o.builder.
		Unstructured().
		ContinueOnError().
		NamespaceParam(o.namespace).DefaultNamespace().
		AllNamespaces(o.allNamespaces)

	if !cmdutil.IsFilenameSliceEmpty(o.FilenameOptions.Filenames, o.FilenameOptions.Kustomize) {
		// handle filename was passed
		r := b.FilenameParam(false, &o.FilenameOptions).
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
		r := b.Flatten().
			ResourceTypeOrNameArgs(false, o.args...).
			Latest().
			Do()
		err := r.Err()
		if err != nil {
			return err
		}

		return r.Visit(func(info *resource.Info, err error) error {
			if err != nil {
				return err
			}

			// Play with the object
			unstructuredObj, ok := info.Object.(*unstructured.Unstructured)
			if !ok {
				return fmt.Errorf("object is not unstructured")
			}
			resultMsg := fmt.Sprintf("Hello %s %s %s", info.Mapping.GroupVersionKind.Kind, info.Name, unstructuredObj.GetCreationTimestamp().Time)

			// check if unstructuredObj has spec.container field and if yes add it to the message
			containers, found, err := unstructured.NestedSlice(unstructuredObj.Object, "spec", "containers")
			if err != nil {
				return fmt.Errorf("error accessing spec.containers: %v", err)
			}

			// If we found containers, print the images
			if found && len(containers) > 0 {
				for _, container := range containers {
					containerMap, ok := container.(map[string]interface{})
					if ok {
						resultMsg += fmt.Sprintf(" %s", containerMap["image"])
					}
				}
			}

			printer, err := o.ToPrinter(resultMsg)
			if err != nil {
				return err
			}
			return printer.PrintObj(info.Object, o.Out)
		})
	}

	return nil
}
