package screenshot

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

func ScreenshotCanvas(rootCtx context.Context, url string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(rootCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var buf []byte

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),

		chromedp.Sleep(3*time.Second),

		chromedp.WaitReady("canvas", chromedp.ByQuery),

		chromedp.Screenshot("canvas", &buf, chromedp.NodeVisible),
	)

	return buf, err
}
