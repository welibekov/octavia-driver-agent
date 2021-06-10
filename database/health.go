package database

import (
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

func updateHealthMonitor(health_monitor_id, pool_id string) {
	load_balancer_id := getLoadbalanceIdFromPoolId(pool_id)
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id)

	updateOperatingStatus(healthMonitor,online,health_monitor_id)
	updateProvisioningStatus(healthMonitor, pendingUpdate, active, health_monitor_id)
	updateProvisioningStatus(pool, pendingUpdate, active, pool_id)
	updateProvisioningStatus(listener, pendingUpdate, active, listener_id)
	updateProvisioningStatus(loadBalancer, pendingUpdate, active, load_balancer_id)
}

func createHealthMonitor(health_monitor_id, pool_id string) {
	load_balancer_id := getLoadbalanceIdFromPoolId(pool_id)
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id)

	updateOperatingStatus(healthMonitor,online,health_monitor_id)
	updateProvisioningStatus(healthMonitor, pendingCreate, active, health_monitor_id)
	updateProvisioningStatus(pool, pendingUpdate, active, pool_id)
	updateProvisioningStatus(listener, pendingUpdate, active, listener_id)
	updateProvisioningStatus(loadBalancer, pendingUpdate, active, load_balancer_id)
}

func deleteHealthMonitor(health_monitor_id, pool_id string) {
	load_balancer_id := getLoadbalanceIdFromPoolId(pool_id)
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id)

	deleteItem(healthMonitor, health_monitor_id)
	updateProvisioningStatus(pool, pendingUpdate, active, pool_id)
	updateProvisioningStatus(listener, pendingUpdate, active, listener_id)
	updateProvisioningStatus(loadBalancer, pendingUpdate, active, load_balancer_id)
}

func UpdateTableHealthMonitor() {
	res, _ := Database.Query("SELECT  project_id, id, operating_status, provisioning_status, pool_id  FROM health_monitor;")
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
		//	updateOperatingStatus(pool,obj.OperatingStatus,pl.Id)
		//}

		// update provisioing_status for pool and and corresponding listener,load_balancer
		if hm.ProvisioningStatus == pendingCreate {
			createHealthMonitor(hm.Id, hm.PoolId)
		} else if hm.ProvisioningStatus == pendingUpdate {
			updateHealthMonitor(hm.Id, hm.PoolId)
		} else if hm.ProvisioningStatus == pendingDelete {
			deleteHealthMonitor(hm.Id, hm.PoolId)
            updateQuota(health_monitor_quota, hm.ProjectId)
		}
	}
}

