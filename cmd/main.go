package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/pytimer/certadm/pkg/kubeadm/v1alpha2"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

func main() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Set("logtostderr", "true")
	// We do not want these flags to show up in --help
	// These MarkHidden calls must be after the lines above
	pflag.CommandLine.MarkHidden("version")
	pflag.CommandLine.MarkHidden("log_flush_frequency")
	pflag.CommandLine.MarkHidden("alsologtostderr")
	pflag.CommandLine.MarkHidden("log_backtrace_at")
	pflag.CommandLine.MarkHidden("log_dir")
	pflag.CommandLine.MarkHidden("logtostderr")
	pflag.CommandLine.MarkHidden("stderrthreshold")
	pflag.CommandLine.MarkHidden("vmodule")
	pflag.CommandLine.MarkHidden("log_file")
	pflag.CommandLine.MarkHidden("log_file_max_size")
	pflag.CommandLine.MarkHidden("skip_headers")
	pflag.CommandLine.MarkHidden("skip_log_headers")

	cmds := &cobra.Command{
		Use:   "certadm",
		Short: "The tool renew or recreate Kubernetes cluster certificates created by kubeadm.",
		Long: dedent.Dedent(`
				┌──────────────────────────────────────────────────────────────────────────────────┐
				│ CERTADM                                                                          │
				│ Renew or recreate Kubernetes cluster certificates created by kubeadm             │
				│                                                                                  │
				└──────────────────────────────────────────────────────────────────────────────────┘
			
			Example usage:

				Renew or recreate Kubernetes cluster certificates created by kubeadm.
				┌──────────────────────────────────────────────────────────┐
				│ Renew certificates:                                      │
				├──────────────────────────────────────────────────────────┤
				│ # certadm renew                                          │
				└──────────────────────────────────────────────────────────┘
				┌──────────────────────────────────────────────────────────┐
				│ Recreate certificates(not implement):                    │
				├──────────────────────────────────────────────────────────┤
				│ # certadm recreate                                       │
				└──────────────────────────────────────────────────────────┘
		`),
	}

	cmds.ResetFlags()
	cmds.AddCommand(NewCmdRenew())

	if err := cmds.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
