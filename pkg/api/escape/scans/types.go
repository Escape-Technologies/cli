package scans

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Scan struct {
	Id            openapi_types.UUID `json:"id"`
	Status        string             `json:"status"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
	FinishedAt    *time.Time         `json:"finishedAt"`
	ProgressRatio float64            `json:"progressRatio"`
	Score     *float32 `json:"score"`
	Initiator string   `json:"initiator"`
}

type IssuesLite struct {
	Id openapi_types.UUID `json:"id"`
	Ignored bool `json:"ignored"`
}

type Test struct {
	Category string `json:"category"`
	SecurityTestUid string `json:"securityTestUid"`
	Meta Meta `json:"meta"`
}

type Meta struct {
	TitleOnFail string `json:"titleOnFail"`
	Type string `json:"type"`
}

type Report struct {
	Id openapi_types.UUID `json:"id"`
	Issues []IssuesLite `json:"issues"`
	Severity string `json:"severity"`
	Ignored bool `json:"ignored"`
	Test Test `json:"test"`
	Meta Meta `json:"meta"`
}

type Issue struct {
	Id openapi_types.UUID `json:"id"`
	Ignored bool `json:"ignored"`
	FirstSeenScanId openapi_types.UUID `json:"firstSeenScanId"`
	LastSeenScanId openapi_types.UUID `json:"lastSeenScanId"`
	Severity string `json:"severity"`
	Risks []Risks `json:"risks"`
}

type Risks struct {
	Id openapi_types.UUID `json:"id"`
	Kind string `json:"kind"`
}

type ScanExchangeArchive struct {
	Archive string `json:"archive"`
}

type ScanEvent struct {
	Id openapi_types.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Description string `json:"description"`
	Level string `json:"level"`
	Title string `json:"title"`
}