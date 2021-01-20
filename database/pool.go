package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/logger"
	"fmt"
)

type PoolTable struct {
	ProjectId				string
	Id						string
	Name					string
	Description				string
	Protocol				string
	LbAlgorithm				string
	OperatingStatus			string
	Enabled					int
	LoadbalancerId			string
	CreatedAt				string
	UpdatedAt				string
	ProvisioningStatus		string
	TlsCertificateId		string
	CaTlsCertificateId		string
	CrlContainerId			string
	TlsEnabled				int
}

func removeDefaultPoolFromSessionPersistence(table, pool_id string, db *sql.DB) {
	del, err := db.Prepare(fmt.Sprintf("DELETE from %s WHERE pool_id=?",table))
	if err != nil {
		logger.Debug(err)
	}
	defer del.Close()
	_, err = del.Exec(pool_id)
	if err != nil {
		logger.Debug(err)
	} else {
		logger.Debug(fmt.Errorf("%s %s: DELETED",table, pool_id))
	}
}

func deletePool(pool_id, load_balancer_id string, db *sql.DB) {
	removeDefaultPoolFromSessionPersistence(sessionPersistence,pool_id,db)
	removeDefaultPoolFromListener(listener,pool_id,load_balancer_id,db)
	deleteItem(pool,pool_id,db)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id,db)
}

func updatePool(pool_id, load_balancer_id string, db *sql.DB) {
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id, db)
	updateProvisioningStatus(pool,pendingUpdate,active,pool_id,db)
	updateProvisioningStatus(listener,pendingUpdate,active,listener_id,db)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id,db)
}

func createPool(pool_id, load_balancer_id string, db *sql.DB) {
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id, db)
	updateProvisioningStatus(pool,pendingCreate,active,pool_id,db)
	updateProvisioningStatus(listener,pendingUpdate,active,listener_id,db)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id,db)
}

func getListenerIdFromLoadbalancerId(load_balancer_id string, db *sql.DB) string {
	res, err := db.Query(fmt.Sprintf("SELECT id FROM listener WHERE load_balancer_id='%s';",load_balancer_id))

	if err != nil {
		logger.Debug(err)
	}
	var ls ListenerTable
	for res.Next() {
		err = res.Scan(
			&ls.Id,
		)
		if err != nil {
			logger.Debug(err)
		}
	}
	return ls.Id
}

func UpdateTablePool(db *sql.DB, obj rabbit.ObjEntity) {
	res, err := db.Query("SELECT  project_id, id, operating_status, provisioning_status, load_balancer_id FROM pool;")
	if err != nil {
		logger.Debug(err)
	}
	var pl PoolTable
	for res.Next() {
		err := res.Scan(
			&pl.ProjectId,
			&pl.Id,
			&pl.OperatingStatus,
			&pl.ProvisioningStatus,
			&pl.LoadbalancerId,
		)
		if err != nil {
			logger.Debug(err)
		}

		// check for operating_status first
		if pl.OperatingStatus != obj.OperatingStatus {
			updateOperatingStatus(pool,obj.OperatingStatus,pl.Id,db)
		}
		// update provisioing_status for pool and and corresponding listener,load_balancer
		if pl.ProvisioningStatus == pendingCreate {
			createPool(pl.Id, pl.LoadbalancerId, db)
		} else if pl.ProvisioningStatus == pendingUpdate {
			updatePool(pl.Id, pl.LoadbalancerId, db)
		} else if pl.ProvisioningStatus == pendingDelete {
			deletePool(pl.Id, pl.LoadbalancerId, db)
		}
	}
}

