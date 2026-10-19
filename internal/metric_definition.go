package internal

import (
	"encoding/json"
	"strings"
)

// Aggregation is the wire representation of a MetricQuery's aggregation
// config. The API models this as a single-element array (0 or 1 items);
// callers work with a plain pointer and wrap/unwrap at the JSON boundary.
type Aggregation struct {
	Time               string   `json:"time"`
	ReduceTo           string   `json:"reduce_to,omitempty"`
	TimeFunctionParams []string `json:"time_function_params,omitempty"`
}

// MetricQuery is the wire representation of one entry in a metric
// definition's `queries` list.
type MetricQuery struct {
	Name        string          `json:"name"`
	Select      []string        `json:"select"`
	From        []string        `json:"from,omitempty"`
	FromExpr    string          `json:"from_expr,omitempty"`
	Where       string          `json:"where"`
	Groups      []string        `json:"groups,omitempty"`
	Aggregation *Aggregation    `json:"-"`
	UnitConfig  json.RawMessage `json:"unit_config,omitempty"`
}

// metricQueryAlias lets MetricQuery customize only the aggregation field's
// JSON shape (a 0-or-1-element array) while reusing struct tags for
// everything else.
type metricQueryAlias struct {
	Name        string          `json:"name"`
	Select      []string        `json:"select"`
	From        []string        `json:"from,omitempty"`
	FromExpr    string          `json:"from_expr,omitempty"`
	Where       string          `json:"where"`
	Groups      []string        `json:"groups,omitempty"`
	Aggregation []Aggregation   `json:"aggregation,omitempty"`
	UnitConfig  json.RawMessage `json:"unit_config,omitempty"`
}

func (q MetricQuery) MarshalJSON() ([]byte, error) {
	alias := metricQueryAlias{
		Name:       q.Name,
		Select:     q.Select,
		From:       q.From,
		FromExpr:   q.FromExpr,
		Where:      q.Where,
		Groups:     q.Groups,
		UnitConfig: q.UnitConfig,
	}
	if q.Aggregation != nil {
		alias.Aggregation = []Aggregation{*q.Aggregation}
	}
	return json.Marshal(alias)
}

func (q *MetricQuery) UnmarshalJSON(data []byte) error {
	var alias metricQueryAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	q.Name = alias.Name
	q.Select = alias.Select
	q.From = alias.From
	q.FromExpr = alias.FromExpr
	q.Where = alias.Where
	q.Groups = alias.Groups
	q.UnitConfig = alias.UnitConfig
	if len(alias.Aggregation) > 0 {
		agg := alias.Aggregation[0]
		q.Aggregation = &agg
	} else {
		q.Aggregation = nil
	}
	return nil
}

// FormulaAggregation is the wire representation of a MetricFormula's
// aggregation config.
type FormulaAggregation struct {
	ReduceTo string `json:"reduce_to,omitempty"`
}

// MetricFormula is the wire representation of one entry in a metric
// definition's `formulas` list.
type MetricFormula struct {
	Name        string              `json:"name"`
	Expression  string              `json:"expression"`
	Aggregation *FormulaAggregation `json:"aggregation,omitempty"`
	UnitConfig  json.RawMessage     `json:"unit_config,omitempty"`
}

// MetricDefinition is the wire representation of the Bronto Metrics API's
// MetricDefinition schema, used for both the create request and the
// create/read responses.
type MetricDefinition struct {
	Id          string          `json:"id"`
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Queries     []MetricQuery   `json:"queries"`
	Formulas    []MetricFormula `json:"formulas,omitempty"`
}

// CreateMetricDefinitionRequest is the body of POST /metrics/definitions.
// There is no update endpoint: any change to a metric definition requires
// deleting and recreating it.
type CreateMetricDefinitionRequest struct {
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Queries     []MetricQuery   `json:"queries"`
	Formulas    []MetricFormula `json:"formulas,omitempty"`
}

// MetricDefinitionsList is the response body of GET /metrics/definitions.
type MetricDefinitionsList struct {
	Definitions []MetricDefinition `json:"definitions"`
}

// DecodeMetricDefinition decodes a single MetricDefinition from a JSON
// response body.
func DecodeMetricDefinition(body string) (MetricDefinition, error) {
	var m MetricDefinition
	err := json.NewDecoder(strings.NewReader(body)).Decode(&m)
	return m, err
}

// DecodeMetricDefinitionsList decodes a {"definitions": [...]} response
// body.
func DecodeMetricDefinitionsList(body string) (MetricDefinitionsList, error) {
	var list MetricDefinitionsList
	err := json.NewDecoder(strings.NewReader(body)).Decode(&list)
	return list, err
}
