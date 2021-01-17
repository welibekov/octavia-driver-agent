package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/logger"
	"fmt"
)

const (
	pendingCreate = "PENDING_CREATE"
	pendingUpdate = "PENDING_UPDATE"
	pendingDelete = "PENDING_DELETE"
	deleted = "DELETED"
	active = "ACTIVE"
	loadBalancer = "load_balancer"
	listener = "listener"
	pool = "pool"
	member = "member"
	vip = "vip"
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

func updateProvisioningStatus(table, old_status, status, id string, db *sql.DB) {
	update, err := db.Prepare(fmt.Sprintf("UPDATE %s SET provisioning_status=? WHERE id=? and provisioning_status=?",table))
	if err != nil {
		logger.Debug(err)
	}
	_, err = update.Exec(status,id,old_status)
	if err != nil {
		logger.Debug(err)
	} else {
		logger.Debug(fmt.Errorf("%s:%s provisioning_status: %s -> %s",table,id,old_status,status))
	}
}

func updateOperatingStatus(table, status, id string, db *sql.DB) {
	update, err := db.Prepare(fmt.Sprintf("UPDATE %s SET operating_status=? WHERE id=?",table))
	if err != nil {
		logger.Debug(err)
	}
	_, err = update.Exec(status,id)
	if err != nil {
		logger.Debug(err)
	} else {
		logger.Debug(fmt.Errorf("%s:%s operating_status: -> %s",table,id,status))
	}
}

func deleteItem(table, id string, db *sql.DB) {
	del, err := db.Prepare(fmt.Sprintf("DELETE from %s WHERE id=?",table))
	if err != nil {
		logger.Debug(err)
	}
	_, err = del.Exec(id)
	if err != nil {
		logger.Debug(err)
	} else {
		logger.Debug(fmt.Errorf("%s %s: DELETED",table, id))
	}
}

func deleteFromVip(table, id string, db *sql.DB) {
	del, err := db.Prepare(fmt.Sprintf("DELETE from %s WHERE load_balancer_id=?",table))
	if err != nil {
		logger.Debug(err)
	}
	_, err = del.Exec(id)
	if err != nil {
		logger.Debug(err)
	} else {
		logger.Debug(fmt.Errorf("%s %s: DELETED",table, id))
	}
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

func findDepsByLoadbalancerId(id, dep string, db *sql.DB) []string {
	deps := []string{}
	res, _ := db.Query(fmt.Sprintf("SELECT id FROM listener WHERE load_balancer_id=?",dep))
	for res.Next() {
		var lb LoadbalancerTable
		err := res.Scan(
			&lb.Id,
		)
		if err != nil {
			logger.Debug(err)
		}
		deps = append(deps, lb.Id)
	}
	return deps
}

func deleteLoadbalancer(load_balancer_id string, db *sql.DB) {
	listeners := findDepsByLoadbalancerId(load_balancer_id, listener, db)
	pools := findDepsByLoadbalancerId(load_balancer_id, pool, db)
	// delete pools first
	for _, pool := range pools {
		deletePool(pool, load_balancer_id, db)
	}
	// delete listeners
	for _, listener := range listeners {
		deleteListener(listener, load_balancer_id, db)
	}
	// delete balancer
	deleteFromVip(vip,load_balancer_id,db)
	deleteItem(loadBalancer,load_balancer_id,db)
}

func deleteListener(listener_id, load_balancer_id string, db *sql.DB) {
	deleteItem(listener,listener_id,db)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id,db)
}

func deletePool(pool_id, load_balancer_id string, db *sql.DB) {
	removeDefaultPoolFromListener(listener,pool_id,load_balancer_id,db)
	deleteItem(pool,pool_id,db)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id,db)
}

// load_balancer table from octavia database
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

		// check for operating_status first
		if lb.OperatingStatus != obj.OperatingStatus {
			updateOperatingStatus(loadBalancer,obj.OperatingStatus,lb.Id,db)
		}
		// if this a new balancer (PENDING_CREATE), update it status to ACTIVE
		if lb.ProvisioningStatus == pendingCreate {
			updateProvisioningStatus(loadBalancer,pendingCreate,active,lb.Id,db)
		} else if lb.ProvisioningStatus == pendingDelete {
			deleteLoadbalancer(lb.Id, db)
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

