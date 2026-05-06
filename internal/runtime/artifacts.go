package runtime

import "os"

type Artifacts struct {
	Manifest Manifest `json:"manifest"`
	State    State    `json:"state"`
}

func NewArtifacts(plan Plan) Artifacts {
	return Artifacts{
		Manifest: NewManifest(plan.ManifestFiles()),
		State:    NewState(plan.Instances),
	}
}

func WriteArtifacts(layout Layout, artifacts Artifacts) error {
	mb, err := artifacts.Manifest.JSON()
	if err != nil {
		return wrapRuntimeStep("serialize manifest", err)
	}
	if err := os.WriteFile(layout.ManifestPath, mb, 0o644); err != nil {
		return wrapRuntimeStep("write manifest", err)
	}

	sb, err := artifacts.State.JSON()
	if err != nil {
		return wrapRuntimeStep("serialize runtime state", err)
	}
	if err := os.WriteFile(layout.RuntimeStatePath, sb, 0o644); err != nil {
		return wrapRuntimeStep("write runtime state", err)
	}
	return nil
}
