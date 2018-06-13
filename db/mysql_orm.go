package db

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/qjpcpu/log"
	"github.com/qjpcpu/log/logging"
	"runtime/debug"
)

func InitOrm(conn_str string) error {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	return orm.RegisterDataBase("default", "mysql", conn_str, 50)
}

func SetDBLog(file_path string) {
	flog, err := logging.NewFileLogWriter(file_path, logging.RotateDaily)
	if err != nil {
		fmt.Println("set db log fail:", err)
		return
	}
	orm.Debug = true
	orm.DebugLog = orm.NewLog(flog)
}

func RegisterModel(models ...interface{}) {
	orm.RegisterModel(models...)
}

func GetOrm() orm.Ormer {
	return orm.NewOrm()
}

type Task func() error

func errHandler(task Task) (err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Criticalf("panic: %v; calltrace:%s", e, string(debug.Stack()))
			err = fmt.Errorf("%v", e)
		}
	}()
	return task()
}

func ExecTransaction(o orm.Ormer, transction ...Task) error {
	if err := o.Begin(); err != nil {
		log.Errorf("DB begin transaction failed: %s", err.Error())
		return err
	}

	for _, task := range transction {
		if task != nil {
			if err := errHandler(task); err != nil {
				if rberr := o.Rollback(); rberr != nil {
					log.Errorf("DB rollback transaction failed: %s", rberr.Error())
				}
				return err
			}
		}
	}

	if err := o.Commit(); err != nil {
		o.Rollback()
		log.Errorf("DB commit transaction failed: %s", err.Error())
		return err
	}
	return nil
}
