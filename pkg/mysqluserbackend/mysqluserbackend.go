package mysqluserbackend

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/cernbox/ocmauthd/pkg"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

type Options struct {
	Hostname string
	Port     int
	Username string
	Password string
	DB       string
	Table    string

	Logger *zap.Logger
}

func New(opt *Options) pkg.UserBackend {

	return &userBackend{
		hostname: opt.Hostname,
		port:     opt.Port,
		username: opt.Username,
		password: opt.Password,
		db:       opt.DB,
		table:    opt.Table,
		logger:   opt.Logger,
		cache:    &sync.Map{},
	}
}

type userBackend struct {
	hostname string
	port     int
	username string
	password string
	db       string
	table    string

	logger *zap.Logger
	cache  *sync.Map
}

// TODO implement caching

func (ub *userBackend) Authenticate(ctx context.Context, token string) (string, string, error) {

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", ub.username, ub.password, ub.hostname, ub.port, ub.db))
	if err != nil {
		ub.logger.Error("CANNOT CONNECT TO MYSQL SERVER", zap.String("HOSTNAME", ub.hostname), zap.Int("PORT", ub.port), zap.String("DB", ub.db))
		return "", "", err
	}
	defer db.Close()

	var user string
	var eos_path string

	query := fmt.Sprintf("SELECT uid_owner, ocm_eos_base_path FROM %s WHERE token=?", ub.table)
	err = db.QueryRow(query, token).Scan(&user, &eos_path)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", errors.New("Invalid path provided")
		}
		ub.logger.Error("CANNOT QUERY STATEMENT")
		return "", "", err
	}

	ub.logger.Info("SHARE AUTHENTICATED", zap.String("user", user), zap.String("eos_path", eos_path))

	return user, eos_path, nil
}

func (ub *userBackend) SetExpiration(ctx context.Context, expiration int64) error {

	return nil
}

func (ub *userBackend) ClearCache(ctx context.Context) {

}
