package main

import (
	"net/http"
	"time"

	"github.com/cernbox/gohub/goconfig"
	"github.com/cernbox/gohub/gologger"
	"github.com/cernbox/ocmauthd/handlers"
	"github.com/cernbox/ocmauthd/pkg/mysqluserbackend"
)

func main() {

	gc := goconfig.New()
	gc.SetConfigName("ocmauthd")
	gc.AddConfigurationPaths("/etc/ocmauthd/")
	gc.Add("tcp-address", "localhost:9991", "tcp address to listen for connections.")
	gc.Add("log-level", "info", "log level to use (debug, info, warn, error).")
	gc.Add("app-log", "stderr", "file to log application information.")
	gc.Add("http-log", "stderr", "file to log HTTP requests.")
	gc.Add("http-read-timeout", 300, "the maximum duration for reading the entire request, including the body.")
	gc.Add("http-write-timeout", 300, "the maximum duration for timing out writes of the response.")
	gc.Add("mysql-hostname", "localhost", "MySQL server hostname.")
	gc.Add("mysql-port", 3306, "MySQL server port.")
	gc.Add("mysql-username", "root", "MySQL server username.")
	gc.Add("mysql-password", "", "MySQL server password.")
	gc.Add("mysql-db", "cbox", "DB name.")
	gc.Add("mysql-table", "oc_share", "Table name.")
	gc.Add("safety-sleep", 5, "Seconds to pause requests on authentication failure.")
	gc.Add("admin-secret", "bar", "secreto to access admin APIs for cache manipulation.")
	gc.BindFlags()
	gc.ReadConfig()

	logger := gologger.New(gc.GetString("log-level"), gc.GetString("app-log"))

	opt := &mysqluserbackend.Options{
		Hostname: gc.GetString("mysql-hostname"),
		Port:     gc.GetInt("mysql-port"),
		Username: gc.GetString("mysql-username"),
		Password: gc.GetString("mysql-password"),
		DB:       gc.GetString("mysql-db"),
		Table:    gc.GetString("mysql-table"),
		Logger:   logger,
	}
	ub := mysqluserbackend.New(opt)

	router := http.NewServeMux()
	authHandler := handlers.BasicAuthOnly(logger, ub, gc.GetInt("safety-sleep"))

	router.Handle("/api/v1/auth", authHandler)

	loggedRouter := gologger.GetLoggedHTTPHandler(gc.GetString("http-log"), router)

	s := http.Server{
		Addr:         gc.GetString("tcp-address"),
		ReadTimeout:  time.Second * time.Duration(gc.GetInt("http-read-timeout")),
		WriteTimeout: time.Second * time.Duration(gc.GetInt("http-write-timeout")),
		Handler:      loggedRouter,
	}

	logger.Info("server is listening at: " + gc.GetString("tcp-address"))
	logger.Error(s.ListenAndServe().Error())

}
