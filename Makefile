build:
	go build -o octavia-driver-agent-go

install:
	cp octavia-driver-agent-go /usr/bin/octavia-driver-agent-go
	cp octavia-driver-agent-go.service /etc/systemd/system/octavia-driver-agent-go.service
	systemctl daemon-reload
	
