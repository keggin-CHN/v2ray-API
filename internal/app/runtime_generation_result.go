package app

import "api-v2ray/internal/xrayruntime"

type RuntimeGenerationResult struct {
	Generated []xrayruntime.GeneratedInstance `json:"generated"`
}
