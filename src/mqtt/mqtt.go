package mqtt

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Mqtt struct {
	clientID string
	client   MQTT.Client
}

type Msg struct {
	Topic string
	Msg   []byte
}

func (m *Mqtt) Init(username, password, host string, port int, tls bool, sendChan <-chan Msg, err chan<- error) error {
	mq := *m
	rand.Seed(time.Now().UnixNano())
	mq.clientID = fmt.Sprintf("connector-%d", rand.Int31())
	er := m.connect(host, m.clientID, username, password, port, tls)
	if er != nil {
		return er
	}
	go m.send(sendChan, err)
	return nil
}

func (m *Mqtt) send(sendChan <-chan Msg, err chan<- error) {
	for {
		msg := <-sendChan

		mqttToken := m.client.Publish(msg.Topic, 0, false, msg.Msg)
		if mqttToken.Wait() && mqttToken.Error() != nil {
			err <- mqttToken.Error()
			return
		}
	}
}

func (m *Mqtt) connect(host, deviceId, user, password string, port int, tlsVerify bool) error {

	clientOpts := MQTT.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%d", host, port)).SetClientID(deviceId).SetCleanSession(true)

	if user != "" {
		clientOpts.SetUsername(user)
		if password != "" {
			clientOpts.SetPassword(password)
		}
	}

	if tlsVerify {
		tlsConfig := &tls.Config{ClientAuth: tls.NoClientCert}
		clientOpts.SetTLSConfig(tlsConfig)
	} else {
		tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
		clientOpts.SetTLSConfig(tlsConfig)
	}

	m.client = MQTT.NewClient(clientOpts)

	if tokenClient := m.client.Connect(); tokenClient.Wait() && tokenClient.Error() != nil {
		return tokenClient.Error()
	}

	return nil
}
