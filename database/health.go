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
			load_balancer_id := getLoadbalanceIdFromPoolId(hm.PoolId, db)
			listener_id := getListenerIdFromLoadbalancerId(load_balancer_id, db)

			updateProvisioningStatus(healthMonitor,pendingCreate,active,hm.Id,db)
			updateProvisioningStatus(pool,pendingUpdate,active,hm.PoolId,db)
			updateProvisioningStatus(listener,pendingUpdate,active,listener_id,db)
			updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id,db)
		}
	}
}

