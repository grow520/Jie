package SCopilot

import (
    "embed"
    "encoding/json"
    "github.com/gin-contrib/pprof"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/yhy0/Jie/conf"
    "github.com/yhy0/Jie/pkg/output"
    "github.com/yhy0/logging"
    "html/template"
    "net/http"
    "runtime"
    "time"
)

/**
   @author yhy
   @since 2023/10/15
   @desc //TODO
**/

//go:embed templates
var templates embed.FS

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // 允许跨域请求
    },
}

// handleWebSocket 使用 websocket 推送数据，同步页面更改
func handleWebSocket(c *gin.Context) {
    host := c.Query("host")
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        logging.Logger.Errorln("Failed to upgrade request to WebSocket:", err)
        return
    }
    defer conn.Close()

    // 数据更改
    for {
        select {
        case <-output.DataUpdated:
            if host != "" {
                data := output.SCopilotMessage[host]
                jsonData, err := json.Marshal(data)
                if err != nil {
                    logging.Logger.Debugln("Failed to marshal data:", err)
                    return
                }
                if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
                    logging.Logger.Debugln("Failed to send message:", err)
                    return
                }
            } else {
                jsonData, err := json.Marshal(output.SCopilotLists)
                if err != nil {
                    logging.Logger.Debugln("Failed to marshal data:", err)
                    return
                }
                if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
                    logging.Logger.Debugln("Failed to send message:", err)
                    return
                }
            }
        }
    }
}

func Init() {
    logging.Logger.Infoln("Start SCopilot web service at :" + conf.GlobalConfig.Passive.WebPort)
    gin.SetMode("release")
    router := gin.Default()

    // 设置模板资源
    router.SetHTMLTemplate(template.Must(template.New("").ParseFS(templates, "templates/*")))

    router.GET("/ws", handleWebSocket)

    // basic 认证
    authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
        conf.GlobalConfig.Passive.WebUser: conf.GlobalConfig.Passive.WebPass,
    }))

    if conf.GlobalConfig.Debug {
        runtime.SetBlockProfileRate(1)     // 开启对阻塞操作的跟踪，block
        runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪，mutex

        pprof.RouteRegister(authorized, "pprof")
    }

    authorized.GET("/", func(c *gin.Context) {
        c.Redirect(302, "/index")
    })

    authorized.GET("/index", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", gin.H{
            "list": output.SCopilotLists,
            "year": time.Now().Year(),
        })
    })

    authorized.GET("/SCopilot", func(c *gin.Context) {
        host := c.Query("host")

        c.HTML(http.StatusOK, "SCopilot.html", gin.H{
            "data":   output.SCopilotMessage[host],
            "ipInfo": output.IPInfoList[output.SCopilotMessage[host].HostNoPort],
            "year":   time.Now().Year(),
        })
    })

    authorized.GET("/config", func(c *gin.Context) {
        c.HTML(http.StatusOK, "config.html", gin.H{
            "plugins":      conf.Plugin,
            "include":      conf.GlobalConfig.Mitmproxy.Include,
            "exclude":      conf.GlobalConfig.Mitmproxy.Exclude,
            "filterSuffix": conf.GlobalConfig.Mitmproxy.FilterSuffix,
            "sqlmap":       conf.GlobalConfig.SqlmapApi,
            "year":         time.Now().Year(),
        })
    })

    // authorized.POST("/config", func(c *gin.Context) {
    //    plugins := c.PostForm("plugin")
    //    include := c.PostForm("include")
    //    exclude := c.PostForm("exclude")
    //    filterSuffix := c.PostForm("filterSuffix")
    //
    //    if plugins != "" {
    //        // 先全部关闭，再根据配置开启对应的，防止配置文件中删除了某个插件，但是程序中还在运行
    //        for k := range conf.Plugin {
    //            conf.Plugin[k] = false
    //        }
    //        for _, plugin := range strings.Split(plugins, ",") {
    //            if plugin != "" {
    //                conf.DefaultPlugins[plugin] = true
    //                mitmproxy.Conf.Plugins = append(mitmproxy.Conf.Plugins, plugin)
    //            }
    //        }
    //        viper.Set("Plugins", mitmproxy.Conf.Plugins)
    //    }
    //
    //    if include != "" {
    //        include = strings.TrimLeft(include, "[")
    //        include = strings.TrimRight(include, "]")
    //        mitmproxy.Conf.Include = strings.Split(include, " ")
    //        viper.Set("Include", mitmproxy.Conf.Include)
    //    }
    //
    //    if exclude != "" {
    //        exclude = strings.TrimLeft(exclude, "[")
    //        exclude = strings.TrimRight(exclude, "]")
    //        mitmproxy.Conf.Exclude = strings.Split(exclude, " ")
    //        viper.Set("Exclude", mitmproxy.Conf.Exclude)
    //    }
    //
    //    if filterSuffix != "" {
    //        filterSuffix = strings.TrimLeft(filterSuffix, "[")
    //        filterSuffix = strings.TrimRight(filterSuffix, "]")
    //        mitmproxy.Conf.FilterSuffix = strings.Split(filterSuffix, " ")
    //        viper.Set("FilterSuffix", mitmproxy.Conf.FilterSuffix)
    //    }
    //
    //    sqlmap := c.PostForm("sqlmap-switch")
    //    sqlmap_api := c.PostForm("sqlmap_api")
    //    username := c.PostForm("username")
    //    password := c.PostForm("password")
    //
    //    if sqlmap == "on" {
    //        conf.DefaultPlugins["sqlmap"] = true
    //        conf.GlobalConfig.SqlmapApi = conf.Sqlmap{
    //            On:       true,
    //            Url:      sqlmap_api,
    //            Username: username,
    //            Password: password,
    //        }
    //    } else {
    //        conf.DefaultPlugins["sqlmap"] = false
    //        conf.GlobalConfig.SqlmapApi = conf.Sqlmap{
    //            On:       false,
    //            Url:      sqlmap_api,
    //            Username: username,
    //            Password: password,
    //        }
    //    }
    //
    //    // 写文件
    //    err := viper.WriteConfigAs(mitmproxy.ConfigFile)
    //    if err != nil {
    //        logging.Logger.Errorln("fail to write 'SCopilot.yaml': %v", err)
    //    }
    //
    //    c.HTML(http.StatusOK, "config.html", gin.H{
    //        "config":       conf.DefaultPlugins,
    //        "include":      mitmproxy.Conf.Include,
    //        "exclude":      mitmproxy.Conf.Exclude,
    //        "filterSuffix": mitmproxy.Conf.FilterSuffix,
    //        "sqlmap":       conf.GlobalConfig.SqlmapApi,
    //        "year":         time.Now().Year(),
    //    })
    // })

    authorized.GET("/clear", func(c *gin.Context) {
        output.SCopilotLists = nil
        output.SCopilotMessage = make(map[string]*output.SCopilotData)
        output.IPInfoList = make(map[string]*output.IPInfo)

        c.HTML(http.StatusOK, "index.html", gin.H{
            "list": output.SCopilotLists,
            "year": time.Now().Year(),
        })
    })

    authorized.GET("/about", func(c *gin.Context) {
        c.HTML(http.StatusOK, "about.html", gin.H{
            "year": time.Now().Year(),
        })
    })

    router.Run(":" + conf.GlobalConfig.Passive.WebPort)
}
