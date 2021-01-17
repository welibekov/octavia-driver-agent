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

// delete load_balancer from vip table
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


func findDepsByLoadbalancerId(id, dep string, db *sql.DB) []string {
	deps := []string{}
	res, _ := db.Query(fmt.Sprintf("SELECT id FROM %s WHERE load_balancer_id=?",dep))
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
	deleteFromVip(vip, load_balancer_id, db)
	deleteItem(loadBalancer, load_balancer_id, db)
}
