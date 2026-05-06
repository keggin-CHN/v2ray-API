package runtime

func (p Plan) Artifacts() Artifacts {
	return NewArtifacts(p)
}
