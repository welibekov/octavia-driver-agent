package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/logger"
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

func UpdateTableMember(db *sql.DB, obj rabbit.ObjEntity) {
	res, _ := db.Query("SELECT  project_id, id, operating_status, provisioning_status, pool_id  FROM member;")
	for res.Next() {
		var mb MemberTable
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
			updateProvisioningStatus(member,pendingCreate,active,mb.Id,db)
			updateProvisioningStatus(pool,pendingUpdate,active,mb.PoolId,db)
		}
	}
}

