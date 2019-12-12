/*l
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
)

var (
	getClusterLong = (`Display a specific cluster or all defined clusters if no name is provided`)

	getClusterExample = fmt.Sprintf(`
# List all the clusters airshipctl knows about
airshipctl config get-cluster

# Display a specific cluster
airshipctl config get-cluster e2e --%v=ephemeral`, config.FlagClusterType)
)

// NewCmdConfigGetCluster returns a Command instance for 'config -Cluster' sub command
func NewCmdConfigGetCluster(rootSettings *environment.AirshipCTLSettings) *cobra.Command {

	theCluster := &config.ClusterOptions{}
	getclustercmd := &cobra.Command{
		Use:     "get-cluster NAME",
		Short:   getClusterLong,
		Example: getClusterExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				theCluster.Name = args[0]
			}
			return runGetCluster(theCluster, cmd.OutOrStdout(), rootSettings)
		},
	}

	gcInitFlags(theCluster, getclustercmd)

	return getclustercmd
}

func gcInitFlags(o *config.ClusterOptions, getclustercmd *cobra.Command) {
	getclustercmd.Flags().StringVar(&o.ClusterType, config.FlagClusterType, "",
		config.FlagClusterType+" for the cluster entry in airshipctl config")
}

// runGetCluster performs the execution of 'config get-cluster' sub command
func runGetCluster(o *config.ClusterOptions, out io.Writer, rootSettings *environment.AirshipCTLSettings) error {
	err := validate(o)
	if err != nil {
		return err
	}

	if o.Name == "" {
		return getClusters(out, rootSettings)
	}
	return getCluster(o.Name, o.ClusterType, out, rootSettings)
}

func getCluster(cName, cType string,
	out io.Writer, rootSettings *environment.AirshipCTLSettings) error {
	airconfig := rootSettings.Config()
	cluster, err := airconfig.GetCluster(cName, cType)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s", cluster.PrettyString())
	return nil
}

func getClusters(out io.Writer, rootSettings *environment.AirshipCTLSettings) error {
	airconfig := rootSettings.Config()
	clusters, err := airconfig.GetClusters()
	if err != nil {
		return err
	}
	if len(clusters) == 0 {
		fmt.Fprintln(out, "No clusters found in the configuration.")
		return nil
	}

	for _, cluster := range clusters {
		fmt.Fprintf(out, "%s\n", cluster.PrettyString())
	}
	return nil
}

func validate(o *config.ClusterOptions) error {
	// Only an error if asking for a specific cluster
	if len(o.Name) == 0 {
		return nil
	}
	return config.ValidClusterType(o.ClusterType)
}
