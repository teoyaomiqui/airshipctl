/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package phase

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/phase"
)

const (
	renderExample = `
# Get all 'initinfra' phase documents containing labels "app=helm" and
# "service=tiller"
airshipctl phase render initinfra -l app=helm,service=tiller

# Get all documents containing labels "app=helm" and "service=tiller"
# and kind 'Deployment'
airshipctl phase render initinfra -l app=helm,service=tiller -k Deployment
`
)

// NewRenderCommand create a new command for document rendering
func NewRenderCommand(cfgFactory config.Factory) *cobra.Command {
	rc := &phase.RenderCommand{}
	renderCmd := &cobra.Command{
		Use:     "render PHASE_NAME",
		Short:   "Render phase documents from model",
		Example: renderExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rc.Options.PhaseID.Name = args[0]
			return rc.RunE(cmd.OutOrStdout())
		},
	}

	addRenderFlags(rc, renderCmd)
	return renderCmd
}

// addRenderFlags adds flags for document render sub-command
func addRenderFlags(rc *phase.RenderCommand, cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVarP(
		&rc.Options.Label,
		"label",
		"l",
		"",
		"filter documents by Labels")

	flags.StringVarP(
		&rc.Options.Annotation,
		"annotation",
		"a",
		"",
		"filter documents by Annotations")

	flags.StringVarP(
		&rc.Options.APIVersion,
		"apiversion",
		"g",
		"",
		"filter documents by API version")

	flags.StringVarP(
		&rc.Options.Kind,
		"kind",
		"k",
		"",
		"filter documents by Kinds")
}
