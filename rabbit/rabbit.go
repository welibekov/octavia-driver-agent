package rabbit

import (
	"github.com/streadway/amqp"
)

const (
	Vmware_nsx__driver_listener = "vmware_nsx__driver_listener"
)

type ObjEntity struct {
    Id string `json:"id"`
    OperatingStatus string `json:"operating_status"`
}

type OsloMsg struct {
    UniqueId string `json:"_unique_id"`
    Args struct {
        Status struct {
            Loadbalancers []ObjEntity `json:"loadbalancers"`
            Pools []ObjEntity `json:"pools"`
            Listeners []ObjEntity `json:"listeners"`
            Members []ObjEntity `json:"members"`
        } `json:"status"`
    } `json:"args"`
    Version string `json:"version"`
    Namespace string `json:"namespace"`
    Method string `json:"method"`
}

type Msg struct {
    OsloMessage string `json:"oslo.message"`
    OsloVersion string `json:"oslo.version"`
}

func Connect(url string) (error, *amqp.Channel) {
	conn, err := amqp.Dial(url)
	if err != nil {
		conn.Close()
		return err, nil
	}

	ch, err := conn.Channel()
	if err != nil {
		ch.Close()
		return err, ch
	}
	return nil, ch
}

