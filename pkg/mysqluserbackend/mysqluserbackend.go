package mysqluserbackend

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"strings"
	"errors"

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

func (ub *userBackend) Authenticate(ctx context.Context, share, token string) error {

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", ub.username, ub.password, ub.hostname, ub.port, ub.db))
	if err != nil {
		ub.logger.Error("CANNOT CONNECT TO MYSQL SERVER", zap.String("HOSTNAME", ub.hostname), zap.Int("PORT", ub.port), zap.String("DB", ub.db))
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(fmt.Sprintf("SELECT share FROM %s WHERE token=?", ub.table))

	if err != nil {
		ub.logger.Error("CANNOT CREATE STATEMENT")
		return err
	}

	rows, err := stmt.Query(token)

	if err != nil {

		if err == sql.ErrNoRows {
			ub.logger.Error("PATH OR TOKEN WRONG")
		} else {
			ub.logger.Error("CANNOT EXECUTE STATEMENT", zap.String("TABLE", ub.table))
		}
		return err
	}

	var returnPath string
	for rows.Next() {
		_ = rows.Scan(returnPath)

		if strings.HasPrefix(share, returnPath) {
			ub.logger.Info("SHARE AUTHENTICATED", zap.String("SHARE", returnPath))
			return nil
		}
	}

	ub.logger.Error("INVALID PATH")
	return errors.New("Invalid path provided")
}

func (ub *userBackend) SetExpiration(ctx context.Context, expiration int64) error {

	return nil
}

func (ub *userBackend) ClearCache(ctx context.Context) {

}
