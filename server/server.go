package server

import (
	//"fmt"
	"database/sql"
	"github.com/streadway/amqp"
	"octavia-driver-agent/rabbit"
	//"octavia-driver-agent/logger"
	"octavia-driver-agent/database"
	"encoding/json"
)

func Run(ch *amqp.Channel, db *sql.DB)  {
	var msg rabbit.Msg
	var o_msg rabbit.OsloMsg
	msgs, _ := ch.Consume(
		rabbit.Vmware_nsx__driver_listener,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	forever := make(chan bool)
	go func() {
		for m := range msgs {
			// becase oslo.message here string
			// we need to unmarshal it also
			json.Unmarshal(m.Body,&msg)
			json.Unmarshal([]byte(msg.OsloMessage),&o_msg)
			// update tables
			for _, loadbalancer := range o_msg.Args.Status.Loadbalancers {
				database.UpdateTableLoadbalancer(db, loadbalancer)
			}
			for _, listener := range o_msg.Args.Status.Listeners {
				database.UpdateTableListener(db, listener)
			}
			for _, pool := range o_msg.Args.Status.Pools {
				database.UpdateTablePool(db, pool)
			}
			for _, member := range o_msg.Args.Status.Members {
				database.UpdateTableMember(db, member)
			}
		}
	}()
	<-forever
}
