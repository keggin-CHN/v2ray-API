package runtime

import "fmt"

func wrapRuntimeStep(step string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", step, err)
}
