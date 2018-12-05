package handlers

import (
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cernbox/ocmauthd/pkg"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func BasicAuthOnly(logger *zap.Logger, userBackend pkg.UserBackend, sleepPause int) http.Handler {
	validBasicAuthsCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "valid_auths_basic",
		Help: "Number of valid authentications using basic authentication.",
	})
	invalidBasicAuthsCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "invalid_auths_basic",
		Help: "Number of valid authentications using basic authentication.",
	})

	prometheus.Register(validBasicAuthsCounter)
	prometheus.Register(invalidBasicAuthsCounter)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, _, ok := r.BasicAuth()
		if !ok {
			invalidBasicAuthsCounter.Inc()
			logger.Info("NO BASIC AUTH PROVIDED")
			time.Sleep(time.Second * time.Duration(sleepPause))
			w.Header().Set("WWW-Authenticate", "Basic Realm='ocmauthd credentials'")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		auth_path := r.Header.Get("auth-path")

		if auth_path == "" || token == "" {
			invalidBasicAuthsCounter.Inc()
			logger.Info("MISSING HEADERS")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		auth_path = filepath.Clean(auth_path)
		path_components := strings.Split(auth_path, "/")

		user, eos_path, err := userBackend.Authenticate(r.Context(), path_components[0], token)
		if err != nil {
			invalidBasicAuthsCounter.Inc()
			logger.Info("WRONG PATH OR TOKEN", zap.String("token", token), zap.String("auth_path", auth_path))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		path_components[0] = eos_path
		full_path := path.Join(path_components...)

		validBasicAuthsCounter.Inc()
		logger.Info("AUTHENTICATION SUCCEEDED", zap.String("PATH", full_path), zap.String("user", user))
		w.Header().Set("user", user)
		w.Header().Set("full_path", full_path)
		w.WriteHeader(http.StatusOK)
	})
}
