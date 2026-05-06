package app

import (
	"api-v2ray/internal/model"
	"api-v2ray/internal/xrayruntime"
)

type BootstrapFlatResult struct {
	Nodes         []model.ProxyNode               `json:"nodes"`
	GeneratedXRAY []xrayruntime.GeneratedInstance `json:"generated_xray"`
}
