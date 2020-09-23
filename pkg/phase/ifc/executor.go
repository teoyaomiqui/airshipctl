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

package ifc

import (
	"io"
	"time"

	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
)

// Executor interface should be implemented by each runner
type Executor interface {
	Run(chan events.Event, RunOptions)
	Render(io.Writer, RenderOptions) error
	Validate() error
}

// RunOptions holds options for run method
type RunOptions struct {
	Debug  bool
	DryRun bool

	Timeout time.Duration
}

// RenderOptions holds options for render method
type RenderOptions struct {
	FilterSelector document.Selector
}

// WaitOptions holds only timeout now, but may be extended in the future
type WaitOptions struct {
	Timeout time.Duration
}

// ExecutorFactory for executor instantiation
type ExecutorFactory func(config ExecutorConfig) (Executor, error)

// ExecutorConfig container to store all executor options
type ExecutorConfig struct {
	PhaseName   string
	ClusterName string

	ClusterMap       clustermap.ClusterMap
	ExecutorDocument document.Document
	AirshipConfig    *config.Config
	Helper           Helper
	KubeConfig       kubeconfig.Interface
	BundleFactory    document.BundleFactoryFunc
}
