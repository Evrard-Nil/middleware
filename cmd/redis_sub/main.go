package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Evrard-Nil/middleware/internal/mqtt_client"

	"github.com/Evrard-Nil/middleware/internal/donneestruct"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
)

var redisCli redis.Conn
var mQTTCli MQTT.Client

func onValueReceived(client MQTT.Client, message MQTT.Message) {
	var sensorData donneestruct.DonneesCapteur
	if err := json.Unmarshal(message.Payload(), &sensorData); err != nil {
		log.Printf("%s", err)
	}
	log.Printf("Received %s value: %s\n", sensorData.Nature, message.Payload())
	key := sensorData.AeroportID + ":" + sensorData.Nature + ":" + strconv.Itoa(sensorData.Date.Year())
	log.Printf("Inserting in key %s", key)
	_, err := redisCli.Do("ZADD", key, strconv.Itoa(int(sensorData.Date.Unix())), message.Payload())
	if err != nil {
		log.Printf("%s", err)
	}
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	redisCli = newRedisClient("redis-10932.c1.us-west-2-2.ec2.cloud.redislabs.com:10932", "uutPD4Eh1qkYtGWxiuYvfXE7Ri5N7oPQ")
	mQTTCli = mqtt_client.Connect("redis_sub")
	mQTTCli.Subscribe("captors/#", 0, onValueReceived)
	defer redisCli.Close()
	defer mQTTCli.Disconnect(250)
	<-c
}

func newRedisClient(addr string, pass string) redis.Conn {
	client, err := redis.Dial("tcp", addr, redis.DialPassword(pass))
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Succesfully connected to Redis at %s\n", addr)
	}
	return client
}
