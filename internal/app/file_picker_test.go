package app

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/eugenioenko/ttt/internal/widgets"
)

func TestFilePickerEntriesOrdersDirectoriesAndFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"zeta", ".hidden-dir", "Alpha"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"zeta.txt", ".hidden.txt", "Alpha.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	gotEntries := filePickerEntries(dir)
	got := make([]string, len(gotEntries))
	for i, entry := range gotEntries {
		got[i] = entry.label
	}
	want := []string{"..", "Alpha", "zeta", ".hidden-dir", "Alpha.txt", "zeta.txt", ".hidden.txt"}
	if !slices.Equal(got, want) {
		t.Fatalf("entries = %v, want %v", got, want)
	}
	for _, entry := range gotEntries[1:4] {
		if !entry.dir {
			t.Errorf("entry %q should be a directory", entry.label)
		}
	}
	for _, entry := range gotEntries[4:] {
		if entry.dir {
			t.Errorf("entry %q should be a file", entry.label)
		}
	}
}

func TestFilterFilePickerEntriesKeepsParent(t *testing.T) {
	entries := []filePickerEntry{
		{label: ".."},
		{label: "app.go"},
		{label: "README.md"},
	}
	got := filterFilePickerEntries(entries, "AP")
	if len(got) != 2 || got[0].label != ".." || got[1].label != "app.go" {
		t.Fatalf("filtered entries = %v, want [.. app.go]", got)
	}
}

func TestFilePickerSyncFromInputPreservesWhitespace(t *testing.T) {
	t.Chdir(t.TempDir())
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("example.txt", nil, 0644); err != nil {
		t.Fatal(err)
	}

	for _, text := range []string{" example.txt", "example.txt ", "example.txt\t", " "} {
		t.Run(text, func(t *testing.T) {
			fp := &filePicker{list: widgets.NewListWidgetFromConfig(widgets.ListConfig{})}
			fp.syncFromInput(text)

			if fp.dir != dir || fp.filter != text {
				t.Fatalf("directory, filter = %q, %q, want %q, %q", fp.dir, fp.filter, dir, text)
			}
			items := fp.list.Config.Items
			if len(items) != 1 || items[0].ID != filepath.Dir(dir) {
				t.Fatalf("filtered listing = %v, want only the parent directory", items)
			}
		})
	}
}

func TestFilePickerSyncFromInputEntersWhitespaceDirectory(t *testing.T) {
	for _, name := range []string{" folder", "folder "} {
		t.Run(name, func(t *testing.T) {
			if runtime.GOOS == "windows" && strings.HasSuffix(name, " ") {
				t.Skip("Windows ordinary paths cannot distinguish filenames with trailing spaces")
			}
			t.Chdir(t.TempDir())
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range []string{"folder", name} {
				if err := os.Mkdir(entry, 0755); err != nil {
					t.Fatal(err)
				}
			}

			fp := &filePicker{list: widgets.NewListWidgetFromConfig(widgets.ListConfig{})}
			fp.syncFromInput(name)

			want := filepath.Join(dir, name)
			if fp.dir != want || fp.filter != "" {
				t.Fatalf("directory, filter = %q, %q, want %q, empty filter", fp.dir, fp.filter, want)
			}
			items := fp.list.Config.Items
			if len(items) != 1 || items[0].ID != dir {
				t.Fatalf("directory listing = %v, want only the parent directory", items)
			}
		})
	}
}
