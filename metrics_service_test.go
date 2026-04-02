package main

import "testing"

func TestNormalizeFeedMetricsURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr bool
	}{
		{
			name:   "share url",
			rawURL: "https://www.xiaohongshu.com/explore/69cb1ee2000000001f001509?source=webshare&xhsshare=pc_web&xsec_token=token&xsec_source=pc_share",
			want:   "https://www.xiaohongshu.com/explore/69cb1ee2000000001f001509?source=webshare&xhsshare=pc_web&xsec_token=token&xsec_source=pc_share",
		},
		{
			name:   "auto add scheme",
			rawURL: "www.xiaohongshu.com/explore/69cb1ee2000000001f001509?xsec_token=token&xsec_source=pc_feed",
			want:   "https://www.xiaohongshu.com/explore/69cb1ee2000000001f001509?xsec_token=token&xsec_source=pc_feed",
		},
		{
			name:    "reject non note path",
			rawURL:  "https://www.xiaohongshu.com/user/profile/123456",
			wantErr: true,
		},
		{
			name:    "reject other host",
			rawURL:  "https://example.com/explore/69cb1ee2000000001f001509",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeFeedMetricsURL(tt.rawURL)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !isFeedMetricsInputError(err) {
					t.Fatalf("expected feedMetricsInputError, got %T", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("normalizeFeedMetricsURL error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("normalizeFeedMetricsURL = %q, want %q", got, tt.want)
			}
		})
	}
}
