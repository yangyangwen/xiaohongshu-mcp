package xiaohongshu

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/errors"
)

// FeedMetrics contains only the four interaction counters needed by the API.
type FeedMetrics struct {
	LikedCount     string `json:"likedCount"`
	CommentCount   string `json:"commentCount"`
	SharedCount    string `json:"sharedCount"`
	CollectedCount string `json:"collectedCount"`
}

var undefinedValuePattern = regexp.MustCompile(`:\s*undefined\s*([,}])`)

// GetFeedMetricsByURL opens the original note URL and extracts interaction counts.
func (f *FeedDetailAction) GetFeedMetricsByURL(ctx context.Context, noteURL string) (*FeedMetrics, error) {
	page := f.page.Context(ctx).Timeout(2 * time.Minute)

	logrus.Infof("Opening feed metrics page: %s", noteURL)

	err := retry.Do(
		func() error {
			page.MustNavigate(noteURL)
			page.MustWaitDOMStable()
			page.MustWait(`() => window.__INITIAL_STATE__ !== undefined`)
			return nil
		},
		retry.Attempts(3),
		retry.Delay(500*time.Millisecond),
		retry.MaxJitter(1000*time.Millisecond),
		retry.OnRetry(func(n uint, err error) {
			logrus.Debugf("Feed metrics navigation retry #%d: %v", n, err)
		}),
	)
	if err != nil {
		return nil, err
	}

	sleepRandom(1000, 1000)

	if err := checkPageAccessible(page); err != nil {
		return nil, err
	}

	return extractFeedMetrics(page)
}

func extractFeedMetrics(page *rod.Page) (*FeedMetrics, error) {
	initialState := page.MustEval(`() => {
		const marker = "window.__INITIAL_STATE__=";
		for (const script of Array.from(document.scripts || [])) {
			const text = script.textContent || "";
			const idx = text.indexOf(marker);
			if (idx >= 0) {
				return text.slice(idx + marker.length);
			}
		}
		return "";
	}`).String()

	if initialState != "" {
		metrics, err := feedMetricsFromInitialState(initialState)
		if err == nil {
			return metrics, nil
		}
		logrus.Warnf("Failed to parse feed metrics from initial state script: %v", err)
	}

	result := page.MustEval(`() => {
		if (window.__INITIAL_STATE__ &&
			window.__INITIAL_STATE__.note &&
			window.__INITIAL_STATE__.note.noteDetailMap) {
			const noteState = window.__INITIAL_STATE__.note;
			const noteId = noteState.currentNoteId
				|| noteState.firstNoteId
				|| Object.keys(noteState.noteDetailMap || {})[0]
				|| "";
			const detail = noteId ? noteState.noteDetailMap[noteId] : null;
			const interactInfo = detail && detail.note ? detail.note.interactInfo : null;
			if (!interactInfo) {
				return "";
			}

			return JSON.stringify({
				likedCount: interactInfo.likedCount || "",
				commentCount: interactInfo.commentCount || "",
				sharedCount: interactInfo.sharedCount || interactInfo.shareCount || "",
				collectedCount: interactInfo.collectedCount || ""
			});
		}
		return "";
	}`).String()

	if result == "" {
		return nil, errors.ErrNoFeedDetail
	}

	var metrics FeedMetrics
	if err := json.Unmarshal([]byte(result), &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal feed metrics: %w", err)
	}

	return &metrics, nil
}

func feedMetricsFromInitialState(initialState string) (*FeedMetrics, error) {
	var payload struct {
		Note struct {
			CurrentNoteID string `json:"currentNoteId"`
			FirstNoteID   string `json:"firstNoteId"`
			NoteDetailMap map[string]struct {
				Note struct {
					InteractInfo InteractInfo `json:"interactInfo"`
				} `json:"note"`
			} `json:"noteDetailMap"`
		} `json:"note"`
	}

	stateJSON := strings.TrimSpace(initialState)
	stateJSON = strings.TrimSuffix(stateJSON, ";")
	stateJSON = undefinedValuePattern.ReplaceAllString(stateJSON, `:null$1`)
	if err := json.Unmarshal([]byte(stateJSON), &payload); err != nil {
		return nil, fmt.Errorf("unmarshal initial state: %w", err)
	}

	noteID := payload.Note.CurrentNoteID
	if noteID == "" {
		noteID = payload.Note.FirstNoteID
	}
	if noteID == "" {
		for id := range payload.Note.NoteDetailMap {
			noteID = id
			break
		}
	}
	if noteID == "" {
		return nil, fmt.Errorf("noteDetailMap is empty")
	}

	noteDetail, exists := payload.Note.NoteDetailMap[noteID]
	if !exists {
		return nil, fmt.Errorf("note %s not found in noteDetailMap", noteID)
	}

	return &FeedMetrics{
		LikedCount:     noteDetail.Note.InteractInfo.LikedCount,
		CommentCount:   noteDetail.Note.InteractInfo.CommentCount,
		SharedCount:    noteDetail.Note.InteractInfo.SharedCount,
		CollectedCount: noteDetail.Note.InteractInfo.CollectedCount,
	}, nil
}
