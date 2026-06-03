// Copyright The Perses Authors
// Licensed under the Apache License, Version 2.0 (the \"License\");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an \"AS IS\" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller_manager

import (
	"github.com/perses/community-mixins/pkg/dashboards"
	panelsGostats "github.com/perses/community-mixins/pkg/panels/gostats"
	panels "github.com/perses/community-mixins/pkg/panels/kubernetes"
	"github.com/perses/community-mixins/pkg/promql"
	"github.com/perses/perses/go-sdk/dashboard"
	panelgroup "github.com/perses/perses/go-sdk/panel-group"
	"github.com/prometheus/prometheus/model/labels"
)

func withCMStatsGroup(datasource string, labelMatcher promql.LabelMatcher) dashboard.Option {
	return dashboard.AddPanelGroup("Controller Manager Status",
		panelgroup.PanelsPerLine(1),
		panelgroup.PanelHeight(8),
		panels.ControllerManagerUpStatus(datasource, labelMatcher),
	)
}

func withCMWorkQueueGroup(datasource string, labelMatcher promql.LabelMatcher) dashboard.Option {
	return dashboard.AddPanelGroup("Work Queue",
		panelgroup.PanelsPerLine(3),
		panelgroup.PanelHeight(8),
		panels.WorkQueueAddRate(datasource, labelMatcher),
		panels.WorkQueueDepth(datasource, labelMatcher),
		panels.WorkQueueLatency(datasource, labelMatcher),
	)
}

func withCMKubeAPIRequestsGroup(datasource string, labelMatcher promql.LabelMatcher) dashboard.Option {
	labelMatchersToUse := []promql.LabelMatcher{
		promql.InstanceVar,
		{
			Name:  "job",
			Value: panels.CONTROLLER_MANAGER_LABEL_VALUE,
			Type:  "=",
		},
	}

	labelMatchersToUse = append(labelMatchersToUse, labelMatcher)

	return dashboard.AddPanelGroup("Kube API Requests",
		panelgroup.PanelsPerLine(3),
		panelgroup.PanelHeight(8),
		panels.KubeAPIRequestRate(datasource, labelMatchersToUse...),
		panels.PostRequestLatency(datasource, labelMatchersToUse...),
		panels.GetRequestLatency(datasource, labelMatchersToUse...),
	)
}

func withCMResources(datasource string, clusterLabelMatcher *labels.Matcher) dashboard.Option {
	// TODO(saswatamcode): Add a way to configure these.
	labelMatchersToUse := []*labels.Matcher{
		promql.InstanceVarV2,
		{
			Name:  "job",
			Value: panels.CONTROLLER_MANAGER_LABEL_VALUE,
			Type:  labels.MatchEqual,
		},
	}

	labelMatchersToUse = append(labelMatchersToUse, clusterLabelMatcher)

	return dashboard.AddPanelGroup("Resource Usage",
		panelgroup.PanelsPerLine(2),
		panelgroup.PanelHeight(8),
		panelsGostats.MemoryUsage(datasource, "instance", labelMatchersToUse...),
		panelsGostats.CPUUsage(datasource, "instance", labelMatchersToUse...),
		panelsGostats.Goroutines(datasource, "instance", labelMatchersToUse...),
		panelsGostats.GarbageCollectionPauseTimeQuantiles(datasource, "instance", labelMatchersToUse...),
	)
}

func BuildControllerManagerOverview(project string, datasource string, clusterLabelName string, variableOverrides ...dashboard.Option) dashboards.DashboardResult {
	defaultVars := []dashboard.Option{
		dashboards.AddClusterVariable(datasource, clusterLabelName, "up{"+panels.GetControllerManagerMatcher()+"}"),
		dashboard.AddVariable("instance",
			listVar.List(
				labelValuesVar.PrometheusLabelValues("instance",
					labelValuesVar.Matchers(
						promql.SetLabelMatchers(
							"up{"+panels.GetControllerManagerMatcher()+"}",
							[]promql.LabelMatcher{dashboards.GetClusterLabelMatcher(clusterLabelName)},
						),
					),
					dashboards.AddVariableDatasource(datasource),
				),
				listVar.DisplayName("instance"),
			),
		),
	}

	clusterLabelMatcher := dashboards.GetClusterLabelMatcher(clusterLabelName)
	clusterLabelMatcherV2 := dashboards.GetClusterLabelMatcherV2(clusterLabelName)

	vars := defaultVars
	if len(variableOverrides) > 0 {
		vars = variableOverrides
	}
	options := append([]dashboard.Option{
		dashboard.ProjectName(project),
		dashboard.Name("Kubernetes / Controller Manager"),
	}, vars...)
	options = append(options,
		withCMStatsGroup(datasource, clusterLabelMatcher),
		withCMWorkQueueGroup(datasource, clusterLabelMatcher),
		withCMKubeAPIRequestsGroup(datasource, clusterLabelMatcher),
		withCMResources(datasource, clusterLabelMatcherV2),
	)
	return dashboards.NewDashboardResult(
		dashboard.New("controller-manager-overview", options...),
	).Component("kubernetes")
}
