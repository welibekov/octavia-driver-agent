package database

import (
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/logger"
	"fmt"
)

type PoolTable struct {
	ProjectId				string
	Id						string
	OperatingStatus			string
	LoadbalancerId			string
	ProvisioningStatus		string
}

func removeDefaultPoolFromSessionPersistence(table, pool_id string) {
	del, err := Database.Prepare(fmt.Sprintf("DELETE from %s WHERE pool_id=?",table))
	if err != nil {
		logger.Debug(err)
	}
	defer del.Close()
	_, err = del.Exec(pool_id)
	if err != nil {
		logger.Debug(err)
	} else {
		logger.Debug(fmt.Errorf("%s %s: DELETED",table, pool_id))
	}
}

func deletePool(pool_id, load_balancer_id string) {
	removeDefaultPoolFromSessionPersistence(sessionPersistence,pool_id)
	removeDefaultPoolFromListener(listener,pool_id,load_balancer_id)
	deleteItem(pool,pool_id)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id)
}

func updatePool(pool_id, load_balancer_id string) {
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id)
	updateProvisioningStatus(pool,pendingUpdate,active,pool_id)
	updateProvisioningStatus(listener,pendingUpdate,active,listener_id)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id)
}

func createPool(pool_id, load_balancer_id string) {
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id)
	updateProvisioningStatus(pool,pendingCreate,active,pool_id)
	updateProvisioningStatus(listener,pendingUpdate,active,listener_id)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id)
}

func getListenerIdFromLoadbalancerId(load_balancer_id string) string {
	res, err := Database.Query(fmt.Sprintf("SELECT id FROM listener WHERE load_balancer_id='%s';",load_balancer_id))

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

func UpdateTablePool(obj rabbit.ObjEntity) {
	res, err := Database.Query("SELECT  project_id, id, operating_status, provisioning_status, load_balancer_id FROM pool;")
	if err != nil {
		logger.Debug(err)
	}
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
			updateOperatingStatus(pool,obj.OperatingStatus,pl.Id)
		}
		// update provisioing_status for pool and and corresponding listener,load_balancer
		if pl.ProvisioningStatus == pendingCreate {
			createPool(pl.Id, pl.LoadbalancerId)
		} else if pl.ProvisioningStatus == pendingUpdate {
			updatePool(pl.Id, pl.LoadbalancerId)
		} else if pl.ProvisioningStatus == pendingDelete {
			deletePool(pl.Id, pl.LoadbalancerId)
		}
	}
}

