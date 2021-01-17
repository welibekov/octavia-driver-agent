package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/logger"
	"fmt"
)

type ListenerTable struct {
	ProjectId						string
	Id								string
	Name							string
	Descriptin						string
	Protocol						string
	ProtocolPort					int
	ConnectionLimit					int
	LoadbalancerId					string
	TlsCertificateId				string
	DefaultPoolId					string
	ProvisioningStatus				string
	OperatingStatus					string
	Enabled							int
	PeerPort						int
	InsertHeaders					string
	CreatedAt						string
	UpdatedAt						string
	TimeoutClientData				int
	TimeoutMemberConnect			int
	TimeoutMemberData				int
	TimeoutTcpInspect				int
	ClientCaTlsCertificateId		string
	ClientAuthentication			string
	ClientCrlContainerId			string
}

func removeDefaultPoolFromListener(table, pool_id, load_balancer_id string, db *sql.DB) {
	update, err := db.Prepare(fmt.Sprintf("UPDATE %s SET default_pool_id=NULL WHERE load_balancer_id=? AND default_pool_id=?",table))
	if err != nil {
		logger.Debug(err)
	}
	_, err = update.Exec(load_balancer_id, pool_id)
	if err != nil {
		logger.Debug(err)
	} else {
		logger.Debug(fmt.Errorf("%s %s default_pool_id deleted",table,pool_id))
	}
}

func deleteListener(listener_id, load_balancer_id string, db *sql.DB) {
	deleteItem(listener,listener_id,db)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id,db)
}

func UpdateTableListener(db *sql.DB, obj rabbit.ObjEntity) {
	res, _ := db.Query("SELECT  project_id, id, operating_status, provisioning_status, load_balancer_id FROM listener;")
	for res.Next() {
		var ls ListenerTable
		err := res.Scan(
			&ls.ProjectId,
			&ls.Id,
			&ls.OperatingStatus,
			&ls.ProvisioningStatus,
			&ls.LoadbalancerId,
		)
		if err != nil {
			logger.Debug(err)
		}

		// check for operating_status first
		if ls.OperatingStatus != obj.OperatingStatus {
			updateOperatingStatus(listener,obj.OperatingStatus,ls.Id,db)
		}
		// update provisioing_status for listener and corresponding load_balancer
		if ls.ProvisioningStatus == pendingCreate {
			updateProvisioningStatus(listener,pendingCreate,active,ls.Id,db)
			updateProvisioningStatus(loadBalancer,pendingUpdate,active,ls.LoadbalancerId,db)
		} else if ls.ProvisioningStatus == pendingDelete {
			deleteListener(ls.Id,ls.LoadbalancerId,db)
			//deleteItem(listener,ls.Id,db)
			//updateProvisioningStatus(loadBalancer,pendingUpdate,active,ls.LoadbalancerId,db)
		}
	}
}

