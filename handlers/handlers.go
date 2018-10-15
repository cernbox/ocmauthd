package handlers

import (
	"net/http"
	"strings"
	"path/filepath"

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

		path := r.Header.Get("auth-path")
		token := r.Header.Get("auth-token")

		if path == "" || token == "" {
			invalidBasicAuthsCounter.Inc()
			logger.Info("MISSING HEADERS")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		path = filepath.Clean(path)

		err := userBackend.Authenticate(r.Context(), path, token)
		if err != nil {
			invalidBasicAuthsCounter.Inc()
			logger.Info("WRONG PATH OR TOKEN")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		path_components := strings.Split(path, "/")
		var user string

		// Assuming EOS username is always a subdirectory of its 1st letter
		// Otherwise we need to remove all possible EOS base paths
		for i, elem := range path_components {
			if len(elem) == 1 {
				if i + 1 < len(path_components) {
					user = path_components[i + 1]
				}
				break
			}
		}

		validBasicAuthsCounter.Inc()
		logger.Info("AUTHENTICATION SUCCEEDED", zap.String("PATH", path))
		w.Header().Set("user", user)
		w.WriteHeader(http.StatusOK)
	})
}
