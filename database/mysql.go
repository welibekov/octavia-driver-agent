package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"octavia-driver-agent/logger"
	"fmt"
)

const (
	pendingCreate = "PENDING_CREATE"
	pendingUpdate = "PENDING_UPDATE"
	pendingDelete = "PENDING_DELETE"
	deleted = "DELETED"
	active = "ACTIVE"
	online = "ONLINE"
	offline = "OFFLINE"
	loadBalancer = "load_balancer"
	listener = "listener"
	pool = "pool"
	member = "member"
	healthMonitor = "health_monitor"
	vip = "vip"
	sessionPersistence = "session_persistence"
)

var Database *sql.DB

func Connect(url string) (error, *sql.DB) {
	db, err := sql.Open("mysql", url)
	if err != nil {
		db.Close()
		return err, db
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return err, db
	}
	return nil, db
}

func updateProvisioningStatus(table, old_status, status, id string) {
	update, err := Database.Prepare(fmt.Sprintf("UPDATE %s SET provisioning_status=? WHERE id=? and provisioning_status=?",table))
	if err != nil {
		logger.Debug(err)
	}
	defer update.Close()
	_, err = update.Exec(status,id,old_status)
	if err != nil {
		logger.Debug(err)
	} else {
		logger.Debug(fmt.Errorf("%s:%s provisioning_status: %s -> %s",table,id,old_status,status))
	}
}

func updateOperatingStatus(table, status, id string) {
	update, err := Database.Prepare(fmt.Sprintf("UPDATE %s SET operating_status=? WHERE id=?",table))
	if err != nil {
		logger.Debug(err)
	}
	defer update.Close()
	_, err = update.Exec(status,id)
	if err != nil {
		logger.Debug(err)
	} else {
		logger.Debug(fmt.Errorf("%s:%s operating_status: -> %s",table,id,status))
	}
}

func deleteItem(table, id string) {
	del, err := Database.Prepare(fmt.Sprintf("DELETE from %s WHERE id=?",table))
	if err != nil {
		logger.Debug(err)
	}
	defer del.Close()
	_, err = del.Exec(id)
	if err != nil {
		logger.Debug(err)
	} else {
		logger.Debug(fmt.Errorf("%s %s: DELETED",table, id))
	}
}

