package engine

import (
    "context"
    "github.com/yhy0/logging"
    "log"
    "sync"
    "time"

    "github.com/chromedp/cdproto/browser"
    "github.com/chromedp/chromedp"
)

type Browser struct {
    Ctx          *context.Context
    Cancel       *context.CancelFunc
    tabs         []*context.Context
    tabCancels   []context.CancelFunc
    ExtraHeaders map[string]interface{}
    lock         sync.Mutex
}

func InitBrowser(proxy string, noHeadless bool) *Browser {
    var bro Browser
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        // 无头模式
        chromedp.Flag("headless", !noHeadless),
        // https://github.com/chromedp/chromedp/issues/997#issuecomment-1030596050
        // incognito mode not used
        // 禁用GPU，不显示GUI
        chromedp.Flag("disable-gpu", true),
        // 取消沙盒模式
        chromedp.Flag("no-sandbox", true),
        // 忽略证书错误
        chromedp.Flag("ignore-certificate-errors", true),

        chromedp.Flag("disable-images", true),
        //
        chromedp.Flag("disable-web-security", true),
        //
        chromedp.Flag("disable-xss-auditor", true),
        //
        chromedp.Flag("disable-setuid-sandbox", true),

        chromedp.Flag("allow-running-insecure-content", true),

        chromedp.Flag("disable-webgl", true),

        chromedp.Flag("disable-popup-blocking", true),

        chromedp.WindowSize(1920, 1080),
    )
    // 设置浏览器代理
    if proxy != "" {
        opts = append(opts, chromedp.ProxyServer(proxy))
    }

    allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
    bctx, _ := chromedp.NewContext(allocCtx,
        chromedp.WithLogf(log.Printf),
    )
    // https://github.com/chromedp/chromedp/issues/824#issuecomment-845664441
    // 如果需要在一个浏览器上创建多个tab，则需要先创建浏览器的上下文，即运行下面的语句
    err := chromedp.Run(bctx)
    if err != nil {
        // not found chrome process need exit
        logging.Logger.Fatal("chromedp run error: ", err.Error())
    }
    bro.Cancel = &cancel
    bro.Ctx = &bctx

    return &bro
}

func ConnectBrowser(wsUrl string, extraHeaders map[string]interface{}) *Browser {
    var bro Browser
    allocCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), wsUrl)
    bctx, _ := chromedp.NewContext(allocCtx,
        chromedp.WithLogf(log.Printf),
    )

    err := chromedp.Run(bctx)
    if err != nil {
        // couldn't connect to the remote browser, need to exit
        logging.Logger.Fatal("chromedp run error: ", err.Error())
    }
    bro.Cancel = &cancel
    bro.Ctx = &bctx
    bro.ExtraHeaders = extraHeaders

    return &bro
}

func (bro *Browser) NewTab(timeout time.Duration) (context.Context, context.CancelFunc) {
    bro.lock.Lock()
    ctx, cancel := chromedp.NewContext(*bro.Ctx)
    // defer cancel()
    tCtx, _ := context.WithTimeout(ctx, timeout)
    bro.tabs = append(bro.tabs, &tCtx)
    bro.tabCancels = append(bro.tabCancels, cancel)
    // defer cancel2()
    bro.lock.Unlock()
    // return bro.Ctx, &cancel
    return tCtx, cancel
}

func (bro *Browser) Close() {
    for _, cancel := range bro.tabCancels {
        cancel()
    }
    for _, ctx := range bro.tabs {
        browser.Close().Do(*ctx)
    }
    browser.Close().Do(*bro.Ctx)
    (*bro.Cancel)()
}
