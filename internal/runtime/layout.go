package runtime

import "path/filepath"

type Layout struct {
	RootDir           string
	XrayDir           string
	SubscriptionDir   string
	ManifestPath      string
	RuntimeStatePath  string
	SubscriptionCache string
}

func NewLayout(rootDir string, subscriptionCache string) Layout {
	return Layout{
		RootDir:           rootDir,
		XrayDir:           filepath.Join(rootDir, "xray"),
		SubscriptionDir:   filepath.Join(rootDir, "subscriptions"),
		ManifestPath:      filepath.Join(rootDir, "manifest.json"),
		RuntimeStatePath:  filepath.Join(rootDir, "runtime-state.json"),
		SubscriptionCache: subscriptionCache,
	}
}
