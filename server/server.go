package server

import (
	"github.com/streadway/amqp"
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/database"
	"encoding/json"
)

func Run(ch *amqp.Channel)  {
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
			for _, member := range o_msg.Args.Status.Members {
				database.UpdateTableMember(member)
			}
			for _, pool := range o_msg.Args.Status.Pools {
				database.UpdateTablePool(pool)
				database.UpdateTableHealthMonitor()
			}
			for _, listener := range o_msg.Args.Status.Listeners {
				database.UpdateTableListener(listener)
			}
			for _, loadbalancer := range o_msg.Args.Status.Loadbalancers {
				database.UpdateTableLoadbalancer(loadbalancer)
			}
		}
	}()
	<-forever
}
