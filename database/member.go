package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/logger"
	"fmt"
)

type MemberTable struct {
	ProjectId			string
	Id					string
	PoolId				string
	SubnetId			string
	IpAddress			string
	ProtocolPort		int
	Weight				int
	OperatingStatus		string
	Enabled				int
	CreatedAt			string
	UpdatedAt			string
	ProvisioningStatus	string
	Name				string
	MonitorAddress		string
	MonitorPort			int
	Backup				int
}

func getLoadbalanceIdFromPoolId(pool_id string, db *sql.DB) string {
	res, err := db.Query(fmt.Sprintf("SELECT load_balancer_id FROM pool WHERE id='%s';",pool_id))
	if err != nil {
		logger.Debug(err)
	}

	var pl PoolTable
	for res.Next() { 
		err = res.Scan(
			&pl.LoadbalancerId,
		)
		if err != nil {
			logger.Debug(err)
		}
	}
	return pl.LoadbalancerId
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

func createMember(member_id, pool_id string, db *sql.DB) {
	load_balancer_id := getLoadbalanceIdFromPoolId(pool_id, db)
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id, db)

	updateProvisioningStatus(member,pendingCreate,active,member_id,db)
	updateProvisioningStatus(pool,pendingUpdate,active,pool_id,db)
	updateProvisioningStatus(listener,pendingUpdate,active,listener_id,db)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id,db)
}

func deleteMember(member_id, pool_id string, db *sql.DB) {
	load_balancer_id := getLoadbalanceIdFromPoolId(pool_id, db)
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id, db)

	deleteItem(member, member_id, db)
	updateProvisioningStatus(pool,pendingUpdate,active,pool_id,db)
	updateProvisioningStatus(listener,pendingUpdate,active,listener_id,db)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id,db)
}

func UpdateTableMember(db *sql.DB, obj rabbit.ObjEntity) {
	res, _ := db.Query("SELECT project_id, id, operating_status, provisioning_status, pool_id FROM member;")
	var mb MemberTable
	for res.Next() {
		err := res.Scan(
			&mb.ProjectId,
			&mb.Id,
			&mb.OperatingStatus,
			&mb.ProvisioningStatus,
			&mb.PoolId,
		)

		if err != nil {
			logger.Debug(err)
		}

		// check for operating_status first
		if mb.OperatingStatus != obj.OperatingStatus {
			updateOperatingStatus(member,obj.OperatingStatus,mb.Id,db)
		}
		if mb.ProvisioningStatus == pendingCreate {
			createMember(mb.Id, mb.PoolId, db)
		} else if mb.ProvisioningStatus == pendingDelete {
			deleteMember(mb.Id, mb.PoolId, db)
		}
	}
}

