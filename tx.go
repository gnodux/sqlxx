package sqlxx

import (
	"database/sql"
	"github.com/cookieY/sqlx"
)

// Tx transaction wrapper
type Tx struct {
	*sqlx.Tx
	m *Factory
}

func (t *Tx) Parse(tplName string, args any) (string, error) {
	if t.m == nil {
		return "", ErrNoManager
	}
	return t.m.ParseSQL(tplName, args)
}
func (t *Tx) NamedExecTpl(tplName string, arg interface{}) (sql.Result, error) {
	query, err := t.Parse(tplName, arg)
	if err != nil {
		return nil, err
	}
	log.Debug("named exec query:", query, arg)
	return t.NamedExec(query, arg)
}
func (t *Tx) ExecTpl(tplName string, args ...interface{}) (sql.Result, error) {
	query, err := t.Parse(tplName, args)
	if err != nil {
		return nil, err
	}
	log.Debug("exec query:", query, args)
	return t.Exec(query, args...)
}

func (t *Tx) GetTpl(dest any, tpl string, args ...any) error {
	query, err := t.Parse(tpl, args)
	if err != nil {
		return err
	}
	log.Debug("get query:", query, args)
	return t.Get(dest, query, args...)
}

func NewTx(tx *sqlx.Tx, m *Factory) *Tx {
	return &Tx{
		Tx: tx,
		m:  m,
	}
}
