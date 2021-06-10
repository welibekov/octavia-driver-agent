package database

import (
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/logger"
	"fmt"
)

type ListenerTable struct {
	ProjectId						string
	Id								string
	LoadbalancerId					string
	DefaultPoolId					string
	ProvisioningStatus				string
	OperatingStatus					string
}

func removeDefaultPoolFromListener(table, pool_id, load_balancer_id string) {
	update, err := Database.Prepare(fmt.Sprintf("UPDATE %s SET default_pool_id=NULL WHERE load_balancer_id=? AND default_pool_id=?",table))
	if err != nil {
		logger.Debug(err)
	}
	defer update.Close()
	_, err = update.Exec(load_balancer_id, pool_id)
	if err != nil {
		logger.Debug(err)
	} else {
		logger.Debug(fmt.Errorf("%s %s default_pool_id DELETED",table,pool_id))
	}
}

func deleteListener(listener_id, load_balancer_id string) {
	deleteItem(listener,listener_id)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id)
}

func updateListener(listener_id, load_balancer_id string) {
	updateProvisioningStatus(listener,pendingUpdate,active,listener_id)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id)
}

func createListener(listener_id, load_balancer_id string) {
	updateProvisioningStatus(listener,pendingCreate,active,listener_id)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id)
}

func UpdateTableListener(obj rabbit.ObjEntity) {
	res, _ := Database.Query("SELECT  project_id, id, operating_status, provisioning_status, load_balancer_id FROM listener;")
	var ls ListenerTable
	for res.Next() {
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
			updateOperatingStatus(listener,obj.OperatingStatus,ls.Id)
		}
		// update provisioing_status for listener and corresponding load_balancer
		if ls.ProvisioningStatus == pendingCreate {
			createListener(ls.Id,ls.LoadbalancerId)
		} else if ls.ProvisioningStatus == pendingUpdate {
			updateListener(ls.Id,ls.LoadbalancerId)
		} else if ls.ProvisioningStatus == pendingDelete {
			deleteListener(ls.Id,ls.LoadbalancerId)
            updateQuota(listener_quota,ls.ProjectId)
		}
	}
}

