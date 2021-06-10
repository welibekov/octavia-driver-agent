package database

import (
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/logger"
	"fmt"
)

type MemberTable struct {
	ProjectId			string
	Id					string
	PoolId				string
	OperatingStatus		string
	ProvisioningStatus	string
}

func getLoadbalanceIdFromPoolId(pool_id string) string {
	res, err := Database.Query(fmt.Sprintf("SELECT load_balancer_id FROM pool WHERE id='%s';",pool_id))
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

func updateMember(member_id, pool_id string) {
	load_balancer_id := getLoadbalanceIdFromPoolId(pool_id)
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id)

	updateProvisioningStatus(member,pendingUpdate,active,member_id)
	updateProvisioningStatus(pool,pendingUpdate,active,pool_id)
	updateProvisioningStatus(listener,pendingUpdate,active,listener_id)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id)
}

func createMember(member_id, pool_id string) {
	load_balancer_id := getLoadbalanceIdFromPoolId(pool_id)
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id)

	updateProvisioningStatus(member,pendingCreate,active,member_id)
	updateProvisioningStatus(pool,pendingUpdate,active,pool_id)
	updateProvisioningStatus(listener,pendingUpdate,active,listener_id)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id)
}

func deleteMember(member_id, pool_id string) {
	load_balancer_id := getLoadbalanceIdFromPoolId(pool_id)
	listener_id := getListenerIdFromLoadbalancerId(load_balancer_id)

	deleteItem(member, member_id)
	updateProvisioningStatus(pool,pendingUpdate,active,pool_id)
	updateProvisioningStatus(listener,pendingUpdate,active,listener_id)
	updateProvisioningStatus(loadBalancer,pendingUpdate,active,load_balancer_id)
}

func UpdateTableMember(obj rabbit.ObjEntity) {
	res, _ := Database.Query("SELECT project_id, id, operating_status, provisioning_status, pool_id FROM member;")
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
			updateOperatingStatus(member,obj.OperatingStatus,mb.Id)
		}
		if mb.ProvisioningStatus == pendingCreate {
			createMember(mb.Id, mb.PoolId)
		} else if mb.ProvisioningStatus == pendingUpdate {
			updateMember(mb.Id, mb.PoolId)
		} else if mb.ProvisioningStatus == pendingDelete {
		    deleteMember(mb.Id, mb.PoolId)
            updateQuota(member_quota, mb.ProjectId)
		}
	}
}

