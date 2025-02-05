package presenters

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/smartcontractkit/chainlink/core/logger"

	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/chainlink/core/services/pipeline"
)

// Corresponds with models.d.ts PipelineRun
type PipelineRunResource struct {
	JAID
	Outputs      []*string                 `json:"outputs"`
	Errors       []*string                 `json:"errors"`
	Inputs       pipeline.JSONSerializable `json:"inputs"`
	TaskRuns     []PipelineTaskRunResource `json:"taskRuns"`
	CreatedAt    time.Time                 `json:"createdAt"`
	FinishedAt   time.Time                 `json:"finishedAt"`
	PipelineSpec PipelineSpec              `json:"pipelineSpec"`
}

// GetName implements the api2go EntityNamer interface
func (r PipelineRunResource) GetName() string {
	return "pipelineRun"
}

func NewPipelineRunResource(pr pipeline.Run) PipelineRunResource {
	var trs []PipelineTaskRunResource
	for i := range pr.PipelineTaskRuns {
		trs = append(trs, NewPipelineTaskRunResource(pr.PipelineTaskRuns[i]))
	}
	// The UI expects all outputs to be strings.
	var outputs []*string
	// Note for async jobs, the output can be nil.
	outs, ok := pr.Outputs.Val.([]interface{})
	if !ok {
		logger.Default.Errorw(fmt.Sprintf("PipelineRunResource: unable to process output type %T", pr.Outputs.Val), "out", pr.Outputs)
	} else if !pr.Outputs.Null && pr.Outputs.Val != nil {
		for _, out := range outs {
			switch v := out.(type) {
			case string:
				s := v
				outputs = append(outputs, &s)
			case map[string]interface{}:
				b, _ := json.Marshal(v)
				bs := string(b)
				outputs = append(outputs, &bs)
			case decimal.Decimal:
				s := v.String()
				outputs = append(outputs, &s)
			case *big.Int:
				s := v.String()
				outputs = append(outputs, &s)
			case float64:
				s := fmt.Sprintf("%f", v)
				outputs = append(outputs, &s)
			case nil:
				outputs = append(outputs, nil)
			default:
				logger.Default.Errorw(fmt.Sprintf("PipelineRunResource: unable to process output type %T", out), "out", out)
			}
		}
	}
	var errors []*string
	for _, err := range pr.Errors {
		if err.Valid {
			s := err.String
			errors = append(errors, &s)
		} else {
			errors = append(errors, nil)
		}
	}
	return PipelineRunResource{
		JAID:         NewJAIDInt64(pr.ID),
		Outputs:      outputs,
		Errors:       errors,
		Inputs:       pr.Inputs,
		TaskRuns:     trs,
		CreatedAt:    pr.CreatedAt,
		FinishedAt:   pr.FinishedAt.ValueOrZero(),
		PipelineSpec: NewPipelineSpec(&pr.PipelineSpec),
	}
}

// Corresponds with models.d.ts PipelineTaskRun
type PipelineTaskRunResource struct {
	Type       pipeline.TaskType `json:"type"`
	CreatedAt  time.Time         `json:"createdAt"`
	FinishedAt time.Time         `json:"finishedAt"`
	Output     *string           `json:"output"`
	Error      *string           `json:"error"`
	DotID      string            `json:"dotId"`
}

// GetName implements the api2go EntityNamer interface
func (r PipelineTaskRunResource) GetName() string {
	return "taskRun"
}

func NewPipelineTaskRunResource(tr pipeline.TaskRun) PipelineTaskRunResource {
	var output *string
	if tr.Output != nil && !tr.Output.Null {
		outputBytes, _ := tr.Output.MarshalJSON()
		outputStr := string(outputBytes)
		output = &outputStr
	}
	var error *string
	if tr.Error.Valid {
		error = &tr.Error.String
	}
	return PipelineTaskRunResource{
		Type:       tr.Type,
		CreatedAt:  tr.CreatedAt,
		FinishedAt: tr.FinishedAt.ValueOrZero(),
		Output:     output,
		Error:      error,
		DotID:      tr.GetDotID(),
	}
}

func NewPipelineRunResources(prs []pipeline.Run) []PipelineRunResource {
	var out []PipelineRunResource

	for _, pr := range prs {
		out = append(out, NewPipelineRunResource(pr))
	}

	return out
}
