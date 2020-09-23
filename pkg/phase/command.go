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
	"io"
	"strings"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

// RunFlags options for phase run command
type RunFlags struct {
	DryRun  bool
	PhaseID ifc.ID
}

// RunCommand phase run command
type RunCommand struct {
	Options RunFlags
	Factory config.Factory
}

// RunE runs the phase
func (c *RunCommand) RunE() error {
	cfg, err := c.Factory()
	if err != nil {
		return err
	}

	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	client := NewClient(helper)

	phase, err := client.PhaseByID(c.Options.PhaseID)
	if err != nil {
		return err
	}
	return phase.Run(ifc.RunOptions{DryRun: c.Options.DryRun})
}

// PlanCommand plan command
type PlanCommand struct {
	Factory config.Factory
	Writer  io.Writer
}

// RunE runs a phase plan command
func (c *PlanCommand) RunE() error {
	cfg, err := c.Factory()
	if err != nil {
		return err
	}

	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	plan, err := helper.Plan()
	if err != nil {
		return err
	}

	return PrintPlan(plan, c.Writer)
}

// RenderFlags holds filters for selector
type RenderFlags struct {
	// Label filters documents by label string
	Label string
	// Annotation filters documents by annotation string
	Annotation string
	// APIVersion filters documents by API group and version
	APIVersion string
	// Kind filters documents by document kind
	Kind    string
	PhaseID ifc.ID
}

// RenderCommand phase render command
type RenderCommand struct {
	Options RenderFlags
	Factory config.Factory
}

// RunE executes render command
func (c *RenderCommand) RunE(out io.Writer) error {
	cfg, err := c.Factory()
	if err != nil {
		return err
	}

	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	client := NewClient(helper)
	phase, err := client.PhaseByID(c.Options.PhaseID)
	if err != nil {
		return err
	}

	groupVersion := strings.Split(c.Options.APIVersion, "/")
	group := ""
	version := groupVersion[0]
	if len(groupVersion) > 1 {
		group = groupVersion[0]
		version = strings.Join(groupVersion[1:], "/")
	}

	sel := document.NewSelector().
		ByLabel(c.Options.Label).
		ByAnnotation(c.Options.Annotation).
		ByGvk(group, version, c.Options.Kind)

	return phase.Render(out, ifc.RenderOptions{FilterSelector: sel})
}
