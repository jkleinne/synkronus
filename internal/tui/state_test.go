package tui

import "testing"

func TestViewStateConstants(t *testing.T) {
	states := []ViewState{
		ViewStorageList, ViewStorageBucketDetail, ViewStorageObjectList,
		ViewStorageObjectDetail, ViewSqlList, ViewSqlInstanceDetail,
		ViewConfigList, ViewConfigEdit,
	}
	seen := make(map[ViewState]bool)
	for i, s := range states {
		if int(s) != i {
			t.Errorf("ViewState %d has value %d, expected %d", i, s, i)
		}
		if seen[s] {
			t.Errorf("duplicate ViewState value: %d", s)
		}
		seen[s] = true
	}
}

func TestOverlayStateConstants(t *testing.T) {
	overlays := []OverlayState{
		OverlayNone, OverlayHelp, OverlayCreateBucket,
		OverlayDeleteConfirm, OverlayConfigAdd, OverlayConfigDelete,
		OverlayDownloadPath,
	}
	seen := make(map[OverlayState]bool)
	for i, o := range overlays {
		if int(o) != i {
			t.Errorf("OverlayState %d has value %d, expected %d", i, o, i)
		}
		if seen[o] {
			t.Errorf("duplicate OverlayState value: %d", o)
		}
		seen[o] = true
	}
}

func TestTabConstants(t *testing.T) {
	tabs := []Tab{TabStorage, TabSql, TabConfig}
	seen := make(map[Tab]bool)
	for i, tab := range tabs {
		if int(tab) != i {
			t.Errorf("Tab %d has value %d, expected %d", i, tab, i)
		}
		if seen[tab] {
			t.Errorf("duplicate Tab value: %d", tab)
		}
		seen[tab] = true
	}
}

func TestTabCount(t *testing.T) {
	if tabCount != 3 {
		t.Errorf("tabCount = %d, expected 3", tabCount)
	}
}

func TestTabNextWraps(t *testing.T) {
	if TabConfig.Next() != TabStorage {
		t.Errorf("TabConfig.Next() = %d, expected TabStorage (%d)", TabConfig.Next(), TabStorage)
	}
}

func TestTabPrevWraps(t *testing.T) {
	if TabStorage.Prev() != TabConfig {
		t.Errorf("TabStorage.Prev() = %d, expected TabConfig (%d)", TabStorage.Prev(), TabConfig)
	}
}

func TestOverlayHasTextInput(t *testing.T) {
	tests := []struct {
		overlay OverlayState
		want    bool
	}{
		{OverlayNone, false},
		{OverlayHelp, false},
		{OverlayCreateBucket, true},
		{OverlayDeleteConfirm, true},
		{OverlayConfigAdd, true},
		{OverlayConfigDelete, false},
		{OverlayDownloadPath, true},
	}
	for _, tt := range tests {
		if got := tt.overlay.hasTextInput(); got != tt.want {
			t.Errorf("OverlayState(%d).hasTextInput() = %v, want %v", tt.overlay, got, tt.want)
		}
	}
}
