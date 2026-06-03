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

package promql

import (
	promqlbuilder "github.com/perses/promql-builder"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

func SetLabelMatchersV2(query parser.Expr, matchers []*labels.Matcher) parser.Expr {
	copy := promqlbuilder.DeepCopyExpr(query)
	for _, l := range matchers {
		copy = LabelsSetPromQLV2(copy, l.Type, l.Name, l.Value)
	}
	return copy
}

func LabelsSetPromQLV2(query parser.Expr, matchType labels.MatchType, name, value string) parser.Expr {
	if name == "" {
		return query
	}

	promqlbuilder.Inspect(query, func(node parser.Node, path []parser.Node) error {
		switch n := node.(type) {
		case *parser.VectorSelector:
			var found bool
			for i, l := range n.LabelMatchers {
				if l.Name == name {
					if value == "" {
						// Remove the matcher
						n.LabelMatchers = append(n.LabelMatchers[:i], n.LabelMatchers[i+1:]...)
					} else {
						n.LabelMatchers[i].Type = matchType
						n.LabelMatchers[i].Value = value
					}
					found = true
					break
				}
			}
			if !found && value != "" {
				n.LabelMatchers = append(n.LabelMatchers, &labels.Matcher{
					Type:  matchType,
					Name:  name,
					Value: value,
				})
			}
		case *parser.BinaryExpr:
			if n.VectorMatching != nil && value == "" {
				// Remove the label from matching clauses
				for i, l := range n.VectorMatching.MatchingLabels {
					if l == name {
						n.VectorMatching.MatchingLabels = append(n.VectorMatching.MatchingLabels[:i], n.VectorMatching.MatchingLabels[i+1:]...)
						break
					}
				}
			}
		case *parser.AggregateExpr:
			if value == "" {
				// Remove the label from aggregation clauses
				for i, l := range n.Grouping {
					if l == name {
						n.Grouping = append(n.Grouping[:i], n.Grouping[i+1:]...)
						break
					}
				}
			}
		}
		return nil
	})

	return query
}
