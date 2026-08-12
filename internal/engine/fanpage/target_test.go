package fanpage

import (
	"testing"

	"zauto/internal/config"
	"zauto/internal/store"
)

func TestResolveFanpageTargetsByIndex(t *testing.T) {
	fps := []store.Fanpage{
		{ID: 1, FBPageID: "615931763399", Name: "615931763399"},
		{ID: 2, FBPageID: "61593132631889", Name: "61593132631889"},
	}
	action := config.Action{Extra: map[string]interface{}{
		"type": "facebook_fanpage_post", "fanpage_index": float64(1),
	}}
	targets, err := SelectFanpages(fps, action)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 || targets[0].FBPageID != "61593132631889" {
		t.Fatalf("targets=%+v", targets)
	}
}

func TestResolveFanpageTargetsAll(t *testing.T) {
	fps := []store.Fanpage{
		{FBPageID: "615931763399"},
		{FBPageID: "61593132631889"},
		{FBPageID: "61592753657118"},
		{FBPageID: "61593073595562"},
	}
	action := config.Action{Extra: map[string]interface{}{
		"type": "facebook_fanpage_post", "fanpage_mode": "all",
	}}
	targets, err := SelectFanpages(fps, action)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 4 {
		t.Fatalf("len=%d want 4", len(targets))
	}
}

func TestResolveFanpageTargetsByID(t *testing.T) {
	fps := []store.Fanpage{{FBPageID: "615931763399"}, {FBPageID: "61593132631889"}}
	action := config.Action{Extra: map[string]interface{}{
		"type": "facebook_fanpage_post", "fanpage_id": "615931763399",
	}}
	targets, err := SelectFanpages(fps, action)
	if err != nil || len(targets) != 1 || targets[0].FBPageID != "615931763399" {
		t.Fatalf("targets=%+v err=%v", targets, err)
	}
}

func TestFanpageLabels(t *testing.T) {
	labels := fanpageLabels(nil, store.Fanpage{Name: "My Page", FBPageID: "123"})
	if len(labels) != 2 {
		t.Fatalf("labels=%v", labels)
	}
}
