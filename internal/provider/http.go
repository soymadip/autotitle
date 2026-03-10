package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mydehq/autotitle/internal/types"
)

// DoWithRetry executes an HTTP request with exponential backoff for 429 errors.
func DoWithRetry(ctx context.Context, client *http.Client, req *http.Request, service string, preRequest func()) (*http.Response, error) {
	const maxRetries = 3
	for i := 0; i <= maxRetries; i++ {
		if preRequest != nil {
			preRequest()
		}
		// Mimic a modern browser to avoid being flagged by WAFs/Gateways
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Pragma", "no-cache")
		req.Header.Set("Sec-Ch-Ua", `"Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"`)
		req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "none")
		req.Header.Set("Sec-Fetch-User", "?1")
		req.Header.Set("Upgrade-Insecure-Requests", "1")

		resp, err := client.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusServiceUnavailable ||
			resp.StatusCode == http.StatusGatewayTimeout {
			_ = resp.Body.Close()
			if i == maxRetries {
				return nil, types.ErrAPIError{
					Service:    service,
					StatusCode: resp.StatusCode,
					Message:    fmt.Sprintf("request failed with status %d after retries", resp.StatusCode),
				}
			}

			// Default wait 2s, or respect Retry-After
			wait := 2 * time.Second
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					wait = time.Duration(seconds) * time.Second
				}
			}

			// Exponential backoff: 2s, 4s, 8s...
			duration := wait * time.Duration(1<<i)

			// Context-aware sleep
			timer := time.NewTimer(duration)
			select {
			case <-timer.C:
				continue
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			}
		}

		return resp, nil
	}
	return nil, fmt.Errorf("request failed after retries")
}
