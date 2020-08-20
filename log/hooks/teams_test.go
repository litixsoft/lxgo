package lxLogHooks_test

/*func TestTeamsHook_Fire(t *testing.T) {

	// logger init
	log := lxLog.InitLogger(os.Stdout, "debug", "fluentd")
	log.AddHook(&lxLogHooks.TeamsHook{
		LogLevels:   []logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel},
		LogsUri:     "https://console.cloud.google.com/logs/viewer?project=litixsoft&customFacets=undefined&limitCustomFacetWidth=true&minLogLevel=0&expandAll=false&timestamp=2020-08-17T15:56:09.716000000Z&advancedFilter=resource.type%3D%22k8s_container%22%0Aresource.labels.cluster_name%3D%22cluster-1%22%0Aresource.labels.namespace_name%3D%22prod%22%0Aresource.labels.container_name%3D%22urogister-backend-prod%22%0Aseverity%3D%22ERROR%22&dateRangeStart=2020-08-17T14:56:10.080Z&dateRangeEnd=2020-08-17T15:56:10.080Z&interval=PT1H",
		HookUri:     "https://outlook.office.com/webhook/60583ba8-c5ce-4430-bbd3-2fa334fae87d@6c6c46b4-fb1d-475e-8011-684739c7ca7e/IncomingWebhook/fc551013af6e4aec959df574b58be45d/249c6ecb-24f4-45ee-a889-436388713e0a",
		ReleaseName: "urogister-backend-prod",
	})

	err1 := errors.New("hier gab es einen schweren fehler")

	log.WithField("client", "urogister").Error(lxLog.GetMessageWithStack(err1.Error()))
	//log.WithField("client", "urogister").Warn("user not found")
	//log.WithField("client", "urogister").Info("connect to mongodb on localhost")

	// Echo instance
	e := echo.New()

	// Middleware, cors, logger, recover
	e.Use(middleware.CORS())
	e.Use(lxLogMiddleware.EchoLogger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", func (c echo.Context) error {

		return echo.NewHTTPError(500, err1)
		//return c.String(http.StatusOK, "Hello, World!")
	})

	// Start server
	e.Logger.Fatal(e.Start(":3000"))

}*/
