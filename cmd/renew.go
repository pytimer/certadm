package main

import "github.com/spf13/cobra"

func NewCmdRenew() *cobra.Command {
	cmd := &cobra.Command{
		Use: "renew",
	}
	return  cmd
}
