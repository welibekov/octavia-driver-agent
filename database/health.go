package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"octavia-driver-agent/logger"
)

type HealthMonitorTable struct {
	Id string
	Name string
	PoolId string
	ProvisioningStatus string
	OperatingStatus string
	ProjectId string
}

func updateHealthMonitor(health_monitor_id, pool_id string, db *sql.DB) {
	load_balancer_id := getLoadbalanceIdFromPoolId(pool_id, db)
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id, db)

	updateOperatingStatus(healthMonitor,online,health_monitor_id,db)
	updateProvisioningStatus(healthMonitor, pendingUpdate, active, health_monitor_id, db)
	updateProvisioningStatus(pool, pendingUpdate, active, pool_id, db)
	updateProvisioningStatus(listener, pendingUpdate, active, listener_id, db)
	updateProvisioningStatus(loadBalancer, pendingUpdate, active, load_balancer_id, db)
}

func createHealthMonitor(health_monitor_id, pool_id string, db *sql.DB) {
	load_balancer_id := getLoadbalanceIdFromPoolId(pool_id, db)
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id, db)

	updateOperatingStatus(healthMonitor,online,health_monitor_id,db)
	updateProvisioningStatus(healthMonitor, pendingCreate, active, health_monitor_id, db)
	updateProvisioningStatus(pool, pendingUpdate, active, pool_id, db)
	updateProvisioningStatus(listener, pendingUpdate, active, listener_id, db)
	updateProvisioningStatus(loadBalancer, pendingUpdate, active, load_balancer_id, db)
}

func deleteHealthMonitor(health_monitor_id, pool_id string, db *sql.DB) {
	load_balancer_id := getLoadbalanceIdFromPoolId(pool_id, db)
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id, db)

	deleteItem(healthMonitor, health_monitor_id, db)
	updateProvisioningStatus(pool, pendingUpdate, active, pool_id, db)
	updateProvisioningStatus(listener, pendingUpdate, active, listener_id, db)
	updateProvisioningStatus(loadBalancer, pendingUpdate, active, load_balancer_id, db)
}

func UpdateTableHealthMonitor(db *sql.DB) {
	res, _ := db.Query("SELECT  project_id, id, operating_status, provisioning_status, pool_id  FROM health_monitor;")
	var hm HealthMonitorTable

	for res.Next() {
		err := res.Scan(
			&hm.ProjectId,
			&hm.Id,
			&hm.OperatingStatus,
			&hm.ProvisioningStatus,
			&hm.PoolId,
		)
		if err != nil {
			logger.Debug(err)
		}

		// check for operating_status first
		//if pl.OperatingStatus != obj.OperatingStatus {
		//	updateOperatingStatus(pool,obj.OperatingStatus,pl.Id,db)
		//}

		// update provisioing_status for pool and and corresponding listener,load_balancer
		if hm.ProvisioningStatus == pendingCreate {
			createHealthMonitor(hm.Id, hm.PoolId, db)
		} else if hm.ProvisioningStatus == pendingUpdate {
			updateHealthMonitor(hm.Id, hm.PoolId, db)
		} else if hm.ProvisioningStatus == pendingDelete {
			deleteHealthMonitor(hm.Id, hm.PoolId, db)
		}
	}
}

