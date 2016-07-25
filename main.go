package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kovetskiy/godocs"
	"github.com/kovetskiy/lorg"
	"github.com/reconquest/colorgful"
	"github.com/seletskiy/hierr"
)

var (
	version = "[manual build]"
	usage   = "nucleus " + version + `

nucleus is a daemon for proxying OAuth requests.

Usage:
    nucleus [options]
    nucleus -h | --help
    nucleus --version

Options:
    -c --config <path>  Use specified configuration file.
                         [default: /etc/nucleus/nucleus.conf]
    --debug             Show debug messages.
    --trace             Show trace messages.
    -h --help           Show this screen.
    --version           Show version.
`
)

var (
	logger    = lorg.NewLog()
	debugMode = false
	traceMode = false
)

func assert(condition bool, message string) {
	if condition == true {
		panic("bug, assertion failure: " + message)
	}
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	args := godocs.MustParse(usage, version, godocs.UsePager)

	logger.SetFormat(
		colorgful.MustApplyDefaultTheme(
			"${time} ${level:[%s]:right:short} ${prefix}%s",
			colorgful.Dark,
		),
	)

	debugMode = args["--debug"].(bool)
	if debugMode {
		logger.SetLevel(lorg.LevelDebug)
	}

	traceMode = args["--trace"].(bool)
	if traceMode {
		logger.SetLevel(lorg.LevelTrace)
	}

	config, err := getConfig(args["--config"].(string))
	if err != nil {
		hierr.Fatalf(
			err, "can't configure daemon",
		)
	}

	app, err := newApp(config)
	if err != nil {
		hierr.Fatalf(
			err, "can't initialize daemon",
		)
	}

	router := gin.New()
	router.Use(getRouterRecovery(), getRouterLogger())

	{
		router.Static("/static/", "static")
		router.LoadHTMLGlob("templates/*.html")

		router.Handle(
			"GET", "/",
			app.handleIndex,
		)

		router.Handle(
			"POST", "/token",
			app.handleGenerateToken,
		)

		router.Handle(
			"GET", "/login/:provider/",
			app.handleLogin,
		)

		router.Handle(
			"GET", "/api/v1/user",
			app.handleUser,
		)
	}

	go watchDatabaseConnection(app.db)

	err = router.RunTLS(
		config.Web.Listen,
		config.Web.TLSCertificate,
		config.Web.TLSKey,
	)
	if err != nil {
		hierr.Fatalf(
			err, "can't run http server at %s", config.Web.Listen,
		)
	}
}

func watchDatabaseConnection(db *database) {
	for {
		time.Sleep(time.Second * 5)

		err := db.ping()
		if err == nil {
			continue
		}

		warningf(
			"database connection has gone away, " +
				"trying to reestablish database connection...",
		)

		err = db.connect()
		if err != nil {
			errorf("can't establish database connection: %s", err)
			continue
		}

		warningf("database connection established")
	}
}
