package xiaohongshu

import (
	"encoding/json"
	"testing"
)

func TestInteractInfoUnmarshalShareCountAlias(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "sharedCount",
			data: `{"liked":false,"likedCount":"1","sharedCount":"2","commentCount":"3","collectedCount":"4","collected":false}`,
			want: "2",
		},
		{
			name: "shareCount",
			data: `{"liked":false,"likedCount":"1","shareCount":"5","commentCount":"3","collectedCount":"4","collected":false}`,
			want: "5",
		},
		{
			name: "prefer sharedCount when both exist",
			data: `{"liked":false,"likedCount":"1","sharedCount":"6","shareCount":"7","commentCount":"3","collectedCount":"4","collected":false}`,
			want: "6",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got InteractInfo
			if err := json.Unmarshal([]byte(tt.data), &got); err != nil {
				t.Fatalf("unmarshal interact info: %v", err)
			}

			if got.SharedCount != tt.want {
				t.Fatalf("SharedCount = %q, want %q", got.SharedCount, tt.want)
			}
		})
	}
}
