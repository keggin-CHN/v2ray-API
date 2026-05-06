package runtime

type FileSet struct {
	ConfigPaths []string `json:"config_paths"`
	LogPaths    []string `json:"log_paths"`
}

func BuildFileSet(rootDir string, specs []LaunchSpec) FileSet {
	out := FileSet{}
	for _, spec := range specs {
		out.ConfigPaths = append(out.ConfigPaths, spec.ConfigPath)
		out.LogPaths = append(out.LogPaths, DefaultLogPath(rootDir, spec.NodeID))
	}
	return out
}
