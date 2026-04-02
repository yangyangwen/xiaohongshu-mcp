package xiaohongshu

import "testing"

func TestFeedMetricsFromInitialState(t *testing.T) {
	t.Parallel()

	initialState := `{
		"global": {
			"firstVisitUrl": undefined
		},
		"note": {
			"currentNoteId": "note_1",
			"noteDetailMap": {
				"note_1": {
					"note": {
						"interactInfo": {
							"likedCount": "11",
							"commentCount": "22",
							"shareCount": "33",
							"collectedCount": "44"
						}
					}
				}
			}
		}
	}`

	got, err := feedMetricsFromInitialState(initialState)
	if err != nil {
		t.Fatalf("feedMetricsFromInitialState error: %v", err)
	}

	if got.LikedCount != "11" {
		t.Fatalf("LikedCount = %q, want %q", got.LikedCount, "11")
	}
	if got.CommentCount != "22" {
		t.Fatalf("CommentCount = %q, want %q", got.CommentCount, "22")
	}
	if got.SharedCount != "33" {
		t.Fatalf("SharedCount = %q, want %q", got.SharedCount, "33")
	}
	if got.CollectedCount != "44" {
		t.Fatalf("CollectedCount = %q, want %q", got.CollectedCount, "44")
	}
}
