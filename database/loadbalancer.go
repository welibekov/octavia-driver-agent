package database

import (
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/logger"
	"fmt"
)

type LoadbalancerTable struct {
	ProjectId			string
	Id					string
	ProvisioningStatus	string
	OperatingStatus		string
}

// load_balancer table from octavia database
func UpdateTableLoadbalancer(obj rabbit.ObjEntity) {
	res, err := Database.Query("SELECT project_id, id, operating_status, provisioning_status FROM load_balancer;")
	if err != nil {
		logger.Debug(err)
	}
	var lb LoadbalancerTable
	for res.Next() {
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
			updateOperatingStatus(loadBalancer,obj.OperatingStatus,lb.Id)
		}
		// if this a new balancer (PENDING_CREATE), update it status to ACTIVE
		if lb.ProvisioningStatus == pendingCreate {
			updateProvisioningStatus(loadBalancer,pendingCreate,active,lb.Id)
		} else if lb.ProvisioningStatus == pendingUpdate {
			updateProvisioningStatus(loadBalancer,pendingUpdate,active,lb.Id)
		} else if lb.ProvisioningStatus == pendingDelete {
			deleteLoadbalancer(lb.Id)
		}
	}
}

// delete load_balancer from vip table
func deleteFromVip(table, id string) {
	del, err := Database.Prepare(fmt.Sprintf("DELETE from %s WHERE load_balancer_id=?",table))
	if err != nil {
		logger.Debug(err)
	}
	defer del.Close()
	_, err = del.Exec(id)
	if err != nil {
		logger.Debug(err)
	} else {
		logger.Debug(fmt.Errorf("%s %s: DELETED",table, id))
	}
}

func findListenersByLoadbalancerId(load_balancer_id string) []string {
	return findDepsByLoadbalancerId(load_balancer_id, listener)
}

func findPoolsByLoadbalancerId(load_balancer_id string) []string {
	return findDepsByLoadbalancerId(load_balancer_id, pool)
}

func findHealthMonitorsByPoolId(pool_id string) []string {
	health_monitors := []string{}
	res, err := Database.Query(fmt.Sprintf("SELECT id FROM health_monitor WHERE pool_id='%s'",pool_id))
	if err != nil {
		logger.Debug(err)
	}
	for res.Next() {
		var hm HealthMonitorTable
		err := res.Scan(
			&hm.Id,
		)
		if err != nil {
			logger.Debug(err)
		}
		health_monitors = append(health_monitors, hm.Id)
	}
	return health_monitors
}

func findMembersByPoolId(pool_id string) []string {
	members := []string{}
	res, err := Database.Query(fmt.Sprintf("SELECT id FROM member WHERE pool_id='%s'",pool_id))
	if err != nil {
		logger.Debug(err)
	}
	for res.Next() {
		var mb MemberTable
		err := res.Scan(
			&mb.Id,
		)
		if err != nil {
			logger.Debug(err)
		}
		members = append(members, mb.Id)
	}
	return members
}

func findDepsByLoadbalancerId(id, dep string) []string {
	deps := []string{}
	res, err := Database.Query(fmt.Sprintf("SELECT id FROM %s WHERE load_balancer_id='%s'",dep,id))
	if err != nil {
		logger.Debug(err)
	}
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

func deleteLoadbalancer(load_balancer_id string) {
	listeners := findListenersByLoadbalancerId(load_balancer_id)
	pools := findPoolsByLoadbalancerId(load_balancer_id)

	// delete pools first
	for _, pool_id := range pools {
		for _, health_monitor_id := range findHealthMonitorsByPoolId(pool_id) {
			deleteHealthMonitor(health_monitor_id, pool_id)
		}
		for _, member_id := range findMembersByPoolId(pool_id) {
			deleteMember(member_id, pool_id)
		}
		deletePool(pool_id, load_balancer_id)
	}

	// delete listeners
	for _, listener_id := range listeners {
		deleteListener(listener_id, load_balancer_id)
	}
	// delete balancer
	deleteFromVip(vip, load_balancer_id)
	deleteItem(loadBalancer, load_balancer_id)
}
