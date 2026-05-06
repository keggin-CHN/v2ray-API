package runtime

func (p Plan) ManifestFiles() []string {
	out := make([]string, 0, len(p.FileSet.ConfigPaths)+len(p.FileSet.LogPaths))
	out = append(out, p.FileSet.ConfigPaths...)
	out = append(out, p.FileSet.LogPaths...)
	return out
}
