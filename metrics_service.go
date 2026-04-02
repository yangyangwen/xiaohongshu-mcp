package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-rod/rod"
	"github.com/xpzouying/xiaohongshu-mcp/xiaohongshu"
)

var allowedFeedMetricsHosts = map[string]struct{}{
	"www.xiaohongshu.com": {},
	"xiaohongshu.com":     {},
	"www.xhslink.com":     {},
	"xhslink.com":         {},
}

type feedMetricsInputError struct {
	message string
}

func (e *feedMetricsInputError) Error() string {
	return e.message
}

func newFeedMetricsInputError(format string, args ...any) error {
	return &feedMetricsInputError{message: fmt.Sprintf(format, args...)}
}

func isFeedMetricsInputError(err error) bool {
	_, ok := err.(*feedMetricsInputError)
	return ok
}

func normalizeFeedMetricsURL(rawURL string) (string, error) {
	noteURL := strings.TrimSpace(rawURL)
	if noteURL == "" {
		return "", newFeedMetricsInputError("缺少 url 参数")
	}

	if !strings.Contains(noteURL, "://") {
		noteURL = "https://" + noteURL
	}

	parsed, err := url.ParseRequestURI(noteURL)
	if err != nil {
		return "", newFeedMetricsInputError("url 格式错误: %v", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", newFeedMetricsInputError("仅支持 http/https 链接")
	}

	host := strings.ToLower(parsed.Hostname())
	if _, ok := allowedFeedMetricsHosts[host]; !ok {
		return "", newFeedMetricsInputError("仅支持小红书笔记链接")
	}

	if host == "www.xiaohongshu.com" || host == "xiaohongshu.com" {
		if !strings.HasPrefix(parsed.Path, "/explore/") {
			return "", newFeedMetricsInputError("仅支持小红书笔记详情链接")
		}
	}

	parsed.Fragment = ""
	return parsed.String(), nil
}

// GetFeedMetricsByURL 根据原始笔记 URL 获取互动数字段。
func (s *XiaohongshuService) GetFeedMetricsByURL(ctx context.Context, rawURL string) (*FeedMetricsResponse, error) {
	noteURL, err := normalizeFeedMetricsURL(rawURL)
	if err != nil {
		return nil, err
	}

	var response *FeedMetricsResponse
	err = withBrowserPage(func(page *rod.Page) error {
		action := xiaohongshu.NewFeedDetailAction(page)
		result, err := action.GetFeedMetricsByURL(ctx, noteURL)
		if err != nil {
			return err
		}

		response = &FeedMetricsResponse{
			LikedCount:     result.LikedCount,
			CommentCount:   result.CommentCount,
			SharedCount:    result.SharedCount,
			CollectedCount: result.CollectedCount,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}
