package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureLayout(t *testing.T) {
	tmp := t.TempDir()
	layout := NewLayout(filepath.Join(tmp, "runtime"), filepath.Join(tmp, "runtime", "subscriptions", "cache.json"))
	if err := EnsureLayout(layout); err != nil { t.Fatalf("ensure layout err: %v", err) }
	for _, p := range []string{layout.RootDir, layout.SubscriptionDir, layout.XrayDir} {
		if st, err := os.Stat(p); err != nil || !st.IsDir() {
			t.Fatalf("expected dir %s, err=%v", p, err)
		}
	}
}
