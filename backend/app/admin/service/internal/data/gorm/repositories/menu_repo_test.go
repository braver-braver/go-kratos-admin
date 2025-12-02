package repositories

import (
	"testing"

	"kratos-admin/app/admin/service/internal/data/gorm/models"
)

func TestMetaFromModel_Defaults(t *testing.T) {
	repo := &MenuRepo{}
	meta := repo.metaFromModel(&models.Menu{})

	if meta == nil {
		t.Fatalf("meta should not be nil")
	}
	if meta.GetHideInMenu() || meta.GetHideInTab() || meta.GetHideInBreadcrumb() || meta.GetKeepAlive() || meta.GetAffixTab() {
		t.Fatalf("expected default flags to be false, got %+v", meta)
	}
	if meta.GetOrder() != 0 {
		t.Fatalf("expected default order 0, got %d", meta.GetOrder())
	}
	if meta.GetIcon() != "" || meta.GetTitle() != "" {
		t.Fatalf("expected empty icon/title, got icon=%s title=%s", meta.GetIcon(), meta.GetTitle())
	}
	if len(meta.GetAuthority()) != 0 {
		t.Fatalf("expected empty authority, got %v", meta.GetAuthority())
	}
}

func TestMetaFromModel_ParseJSON(t *testing.T) {
	repo := &MenuRepo{}
	jsonMeta := `{
        "authority": ["admin", "user"],
        "hideInMenu": true,
        "hideInTab": true,
        "hideInBreadcrumb": true,
        "icon": "lucide:area-chart",
        "keepAlive": true,
        "order": 5,
        "title": "page.dashboard.analytics",
        "affixTab": true
    }`

	meta := repo.metaFromModel(&models.Menu{Meta: &jsonMeta})

	if meta == nil {
		t.Fatalf("meta should not be nil")
	}
	if !meta.GetHideInMenu() || !meta.GetHideInTab() || !meta.GetHideInBreadcrumb() || !meta.GetKeepAlive() || !meta.GetAffixTab() {
		t.Fatalf("expected flags to be true after parse, got %+v", meta)
	}
	if meta.GetOrder() != 5 {
		t.Fatalf("expected order=5, got %d", meta.GetOrder())
	}
	if meta.GetIcon() != "lucide:area-chart" {
		t.Fatalf("unexpected icon: %s", meta.GetIcon())
	}
	if meta.GetTitle() != "page.dashboard.analytics" {
		t.Fatalf("unexpected title: %s", meta.GetTitle())
	}
	wantAuth := []string{"admin", "user"}
	gotAuth := meta.GetAuthority()
	if len(gotAuth) != len(wantAuth) {
		t.Fatalf("authority length mismatch, got %v", gotAuth)
	}
	for i, v := range wantAuth {
		if gotAuth[i] != v {
			t.Fatalf("authority[%d] mismatch: got %s want %s", i, gotAuth[i], v)
		}
	}
}

// Ensure metaFromModel tolerates partial/malformed JSON and still returns defaults for missing fields.
func TestMetaFromModel_Partial(t *testing.T) {
	repo := &MenuRepo{}
	jsonMeta := `{"order":"7","icon":"lucide:test"}`
	meta := repo.metaFromModel(&models.Menu{Meta: &jsonMeta})

	if meta.GetOrder() != 7 {
		t.Fatalf("expected order=7, got %d", meta.GetOrder())
	}
	if meta.GetIcon() != "lucide:test" {
		t.Fatalf("unexpected icon: %s", meta.GetIcon())
	}
	// defaults preserved
	if meta.GetKeepAlive() || meta.GetHideInMenu() || meta.GetHideInTab() || meta.GetHideInBreadcrumb() || meta.GetAffixTab() {
		t.Fatalf("expected defaults for boolean fields, got %+v", meta)
	}
	if len(meta.GetAuthority()) != 0 {
		t.Fatalf("expected empty authority for partial meta, got %v", meta.GetAuthority())
	}
}
