package handlers

import (
	"net/http"
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
			w.Header().Set("WWW-Authenticate", "Basic Realm='ocmauthd credentials'")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if token == "" {
			invalidBasicAuthsCounter.Inc()
			logger.Info("MISSING HEADERS")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		user, eosPath, err := userBackend.Authenticate(r.Context(), token)
		if err != nil {
			invalidBasicAuthsCounter.Inc()
			logger.Info("WRONG PATH OR TOKEN", zap.String("token", token))
			time.Sleep(time.Second * time.Duration(sleepPause))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		validBasicAuthsCounter.Inc()
		logger.Info("AUTHENTICATION SUCCEEDED", zap.String("fullpath", eosPath), zap.String("user", user))
		w.Header().Set("user", user)
		w.Header().Set("eos_path", eosPath)
		w.WriteHeader(http.StatusOK)
	})
}
