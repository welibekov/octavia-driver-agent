package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/logger"
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

func deletePool(pool_id, load_balancer_id string, db *sql.DB) {
	removeDefaultPoolFromListener(listener,pool_id,load_balancer_id,db)
	deleteItem(pool,pool_id,db)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id,db)
}

func UpdateTablePool(db *sql.DB, obj rabbit.ObjEntity) {
	res, _ := db.Query("SELECT  project_id, id, operating_status, provisioning_status, load_balancer_id FROM pool;")
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
			updateProvisioningStatus(pool,pendingCreate,active,pl.Id,db)
			updateProvisioningStatus(listener,pendingUpdate,active,pl.LoadbalancerId,db)
			updateProvisioningStatus(loadBalancer,pendingUpdate,active,pl.LoadbalancerId,db)
		} else if pl.ProvisioningStatus == pendingDelete {
			deletePool(pl.Id, pl.LoadbalancerId, db)
			//removeDefaultPoolFromListener(listener,pl.Id,pl.LoadbalancerId,db)
			//deleteItem(pool,pl.Id,db)
			//updateProvisioningStatus(loadBalancer,pendingUpdate,active,pl.LoadbalancerId,db)
		}
	}
}

