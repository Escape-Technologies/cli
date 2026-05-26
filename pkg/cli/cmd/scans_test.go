package cmd

import (
	"io"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestResolveKinds_Defaults(t *testing.T) {
	cmd := newTestCommand(t)

	got := resolveScanKinds(cmd)

	if !reflect.DeepEqual(got, defaultScanKinds) {
		t.Fatalf("expected default scan kinds %v, got %v", defaultScanKinds, got)
	}
}

func TestResolveKinds_AllKinds(t *testing.T) {
	cmd := newTestCommand(t)
	if err := cmd.Flags().Set("all-kinds", "true"); err != nil {
		t.Fatalf("failed to set all-kinds flag: %v", err)
	}

	got := resolveScanKinds(cmd)

	if got != nil {
		t.Fatalf("expected nil kinds for all-kinds, got %v", got)
	}
	if filter := scanKindsFilter(got); filter != nil {
		t.Fatalf("expected nil filter for all-kinds, got %v", *filter)
	}
}

func TestResolveKinds_KindOverrides(t *testing.T) {
	cmd := newTestCommand(t)
	if err := cmd.Flags().Set("all-kinds", "true"); err != nil {
		t.Fatalf("failed to set all-kinds flag: %v", err)
	}
	if err := cmd.Flags().Set("kind", "ASM_REST"); err != nil {
		t.Fatalf("failed to set kind flag: %v", err)
	}

	got := resolveScanKinds(cmd)
	want := []string{"ASM_REST"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected explicit kinds %v, got %v", want, got)
	}
}

func newTestCommand(t *testing.T) *cobra.Command {
	t.Helper()

	prevKinds := scanKinds
	prevAllKinds := scanListAllKinds
	t.Cleanup(func() {
		scanKinds = prevKinds
		scanListAllKinds = prevAllKinds
	})

	scanKinds = []string{}
	scanListAllKinds = false

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)
	cmd.Flags().StringSliceVarP(&scanKinds, "kind", "k", []string{}, "")
	cmd.Flags().BoolVar(&scanListAllKinds, "all-kinds", false, "")
	return cmd
}
