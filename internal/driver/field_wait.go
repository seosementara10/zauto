package driver

import (
	"fmt"
	"time"

	"zauto/internal/ui"
)

func (d *AdbUI) tapAndType(r *ui.Resolved, value string) error {
	x, y := r.EditTapPoint()
	if err := d.client.Tap(x, y); err != nil {
		return err
	}
	ref := r.Element
	if err := d.waitFieldReady(ref, 3*time.Second); err != nil {
		return err
	}
	if err := d.client.InputText(value); err != nil {
		return err
	}
	// Password EditText values are masked in UI dumps — hierarchy verify always fails.
	if ref.Password {
		return nil
	}
	if err := d.waitFieldContains(ref, value, 5*time.Second); err != nil {
		return fmt.Errorf("field verify after input: %w", err)
	}
	return nil
}

func (d *AdbUI) waitFieldReady(ref ui.Element, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snap := d.read(false)
		if field := ui.FindEditAtBounds(snap, ref); field != nil {
			if field.Focused {
				return nil
			}
		}
		time.Sleep(d.poll)
	}
	// Some devices omit focused="true"; proceed if the edit is still present.
	snap := d.read(false)
	if ui.FindEditAtBounds(snap, ref) != nil {
		return nil
	}
	return fmt.Errorf("field not ready after tap at (%d,%d)", ref.Bounds[0], ref.Bounds[1])
}

func (d *AdbUI) waitFieldContains(ref ui.Element, value string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snap := d.read(false)
		if field := ui.FindEditAtBounds(snap, ref); field != nil && ui.FieldTextContains(*field, value) {
			return nil
		}
		time.Sleep(d.poll)
	}
	return fmt.Errorf("field value not visible: want prefix %q", trimPreview(value))
}

func trimPreview(value string) string {
	if len(value) <= 24 {
		return value
	}
	return value[:24] + "..."
}
