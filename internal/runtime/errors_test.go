package runtime

import (
	"errors"
	"strings"
	"testing"
)

func TestWrapRuntimeStep(t *testing.T) {
	err := wrapRuntimeStep("write manifest", errors.New("boom"))
	if err == nil { t.Fatalf("expected error") }
	if !strings.Contains(err.Error(), "write manifest") { t.Fatalf("unexpected error: %v", err) }
}
