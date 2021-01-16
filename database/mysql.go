package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/logger"
	"fmt"
)

type LoadbalancerTable struct {
	ProjectId			string
	Id					string
	Name				string
	Description			sql.NullString
	ProvisioningStatus	string
	OperatingStatus		string
	Enabled				int
	Topology			string
	ServerGroupId		sql.NullString
	CreatedAt			string
	UpdatedAt			sql.NullString
	Provider			string
	FlavorId			sql.NullString
}

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

func Connect(url string) (error, *sql.DB) {
	//db, err := sql.Open("mysql", "octavia:octavia@tcp(127.0.0.1)/octavia")
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

func UpdateTableLoadbalancer(db *sql.DB, obj rabbit.ObjEntity) {
	res, _ := db.Query("SELECT  project_id, id, operating_status, provisioning_status FROM load_balancer;")
	for res.Next() {
		var lb LoadbalancerTable
		err := res.Scan(
			&lb.ProjectId,
			&lb.Id,
			&lb.OperatingStatus,
			&lb.ProvisioningStatus,
		)
		if err != nil {
			logger.Debug(err)
		}

		if obj.OperatingStatus == "ONLINE" && ( lb.ProvisioningStatus == "PENDING_CREATE" || lb.ProvisioningStatus == "PENDING_UPDATE" ) {
			update, err := db.Prepare("UPDATE load_balancer SET operating_status=?, provisioning_status=? WHERE id=?")
			if err != nil {
				logger.Debug(err)
			}
			_, err = update.Exec(obj.OperatingStatus,"ACTIVE",lb.Id)
			if err != nil {
				logger.Debug(err)
			} else {
				logger.Debug(fmt.Errorf("load_balancer:%s updated from provisioing_status:%s -> ACTIVE,operating_status:%s -> %s",
					lb.Id,lb.ProvisioningStatus,lb.OperatingStatus,obj.OperatingStatus))
			}
		} else if lb.ProvisioningStatus == "PENDING_DELETE" {
			del, err := db.Prepare("UPDATE load_balancer SET operating_status=?, provisioning_status=? WHERE id=?")
			if err != nil {
				logger.Debug(err)
			}
			_, err = del.Exec("OFFLINE","DELETED",lb.Id)
			if err != nil {
				logger.Debug(err)
			}
		}
	}
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

		if ( obj.OperatingStatus == "ONLINE" || obj.OperatingStatus == "OFFLINE" ) && ( ls.ProvisioningStatus == "PENDING_CREATE" || ls.ProvisioningStatus == "PENDING_UPDATE" ) {
			update, err := db.Prepare("UPDATE listener SET operating_status=?, provisioning_status=? WHERE id=? and load_balancer_id=?")
			if err != nil {
				logger.Debug(err)
			}
			_, err = update.Exec(obj.OperatingStatus,"ACTIVE",ls.Id,ls.LoadbalancerId)
			if err != nil {
				logger.Debug(err)
			}
		} else if ls.ProvisioningStatus == "PENDING_DELETE" {
			del, err := db.Prepare("UPDATE listener SET operating_status=?, provisioning_status=? WHERE id=? and load_balancer_id=?")
			if err != nil {
				logger.Debug(err)
			}
			del.Exec("OFFLINE","DELETED",ls.Id,ls.LoadbalancerId)
		}
	}
}

func UpdateTablePool(db *sql.DB, obj rabbit.ObjEntity) {
	res, _ := db.Query("SELECT  project_id, id, operating_status, provisioning_status, load_balancer_id FROM pool;")
	for res.Next() {
		var pl PoolTable
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

		if ( obj.OperatingStatus == "ONLINE" || obj.OperatingStatus == "OFFLINE" ) && ( pl.ProvisioningStatus == "PENDING_CREATE" || pl.ProvisioningStatus == "PENDING_UPDATE" ) {
			update, err := db.Prepare("UPDATE pool SET operating_status=?, provisioning_status=? WHERE id=? and load_balancer_id=?")
			if err != nil {
				logger.Debug(err)
			}
			_, err = update.Exec(obj.OperatingStatus,"ACTIVE",pl.Id,pl.LoadbalancerId)
			if err != nil {
				logger.Debug(err)
			}
		} else if pl.ProvisioningStatus == "PENDING_DELETE" {
			del, err := db.Prepare("UPDATE pool SET operating_status=?, provisioning_status=? WHERE id=? and load_balancer_id=?")
			if err != nil {
				logger.Debug(err)
			}
			_, err = del.Exec("OFFLINE","DELETED",pl.Id,pl.LoadbalancerId)
			if err != nil {
				logger.Debug(err)
			}
		}
	}
}

func UpdateTableMember(db *sql.DB, obj rabbit.ObjEntity) {
	res, _ := db.Query("SELECT  project_id, id, operating_status, provisioning_status, pool_id  FROM member;")
	for res.Next() {
		var mb MemberTable
		res.Scan(
			&mb.ProjectId,
			&mb.Id,
			&mb.OperatingStatus,
			&mb.ProvisioningStatus,
			&mb.PoolId,
		)
		if ( obj.OperatingStatus == "ONLINE" || obj.OperatingStatus == "OFFLINE" ) && ( mb.ProvisioningStatus == "PENDING_CREATE" || mb.ProvisioningStatus == "PENDING_UPDATE" ) {
			update, err := db.Prepare("UPDATE member SET operating_status=?, provisioning_status=? WHERE id=? and pool_id=?")
			if err != nil {
				logger.Debug(err)
			}
			_, err = update.Exec(obj.OperatingStatus,"ACTIVE",mb.Id,mb.PoolId)
			if err != nil {
				logger.Debug(err)
			}
		} else if mb.ProvisioningStatus == "PENDING_DELETE" {
			del, err := db.Prepare("UPDATE member SET operating_status=?, provisioning_status=? WHERE id=? and pool_id=?")
			if err != nil {
				logger.Debug(err)
			}
			_, err = del.Exec("OFFLINE","DELETED",mb.Id,mb.PoolId)
			if err != nil {
				logger.Debug(err)
			}
		}
	}
}

