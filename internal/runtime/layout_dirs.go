package runtime

import "os"

func EnsureLayout(layout Layout) error {
	if err := os.MkdirAll(layout.RootDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(layout.SubscriptionDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(layout.XrayDir, 0o755); err != nil {
		return err
	}
	return nil
}
