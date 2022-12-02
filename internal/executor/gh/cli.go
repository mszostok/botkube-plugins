package gh

import (
	"github.com/spf13/cobra"
)

// CreateIssueOptions holds CLI configuration for creating GitHub issue.
type CreateIssueOptions struct {
	For       string
	Namespace string
	Reason    string
}

// NewRoot returns a root cobra.Command for the whole gh CLI.
func NewRoot(opts *CreateIssueOptions) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "gh",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	create := &cobra.Command{
		Use: "create",
	}
	create.AddCommand(newCreateIssueCmd(opts))

	rootCmd.AddCommand(create)

	return rootCmd
}

func newCreateIssueCmd(opts *CreateIssueOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use: "issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "default", "Kubernetes namespace")
	cmd.Flags().StringVar(&opts.Reason, "reason", "", "Reason of a given issue")
	cmd.Flags().StringVar(&opts.For, "for", "", "Kubernetes object, syntax {kind}/{name}")

	return cmd
}
