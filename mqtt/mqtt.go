package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

/**
* Author: joker
* TODO: test
* Date: 2022/11/17
* Time: 下午3:35
**/

const (
	ADDRESS   = "tcp://192.168.78.141:1883"
	USER_NAME = "admin"
	PASSWORD  = "public"
	TOPIC     = "connect-custom"
)

var (
	MqttClient mqtt.Client
)

const (
	QoS0 = 0 // 至多一次
	QoS1 = 1 // 至少一次
	QoS2 = 2 // 确保只有一次
)

func main() {
	initMqtt()
}

// initMqtt
/**
 *  @Description: 初始化MQTT
 */
func initMqtt() {
	opts := mqtt.NewClientOptions()
	// 添加代理
	opts.AddBroker(ADDRESS)
	// 设置用户名
	opts.SetUsername(USER_NAME)
	// 设置密码
	opts.SetPassword(PASSWORD)
	// 使用连接信息进行连接
	MqttClient = mqtt.NewClient(opts)
	if token := MqttClient.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("订阅 MQTT 失败")
		panic(token.Error())
	}

	fmt.Println("开始推送")

	//for true {
	//	subscribe()
	//}

	type student struct {
		ID int `json:"id"`
	}

	for i := 0; i < 10; i++ {
		s := student{
			ID: i,
		}
		sb, _ := json.Marshal(s)
		publish(string(sb))
	}

	//go func() {
	//	for i := 0; i < 10; i++ {
	//		s := student{
	//			ID: i,
	//		}
	//		sb, _ := json.Marshal(s)
	//		publish(string(sb))
	//	}
	//}()

	select {}
}

// publish
/**
 *  @Description: 发布消息
 *  @param msg
 */
func publish(msg string) {
	MqttClient.Publish(TOPIC, QoS2, true, msg)
	fmt.Println("push success")
}

// subscribe
/**
 *  @Description: 订阅
 */
func subscribe() {
	MqttClient.Subscribe(TOPIC, QoS2, subCallBackFunc)
}

// subCallBackFunc
/**
 *  @Description: 回调函数
 *  @param client
 *  @param msg
 */

func subCallBackFunc(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("订阅: 当前话题是 [%s]; 信息是 [%s] \n", msg.Topic(), msg)
}
