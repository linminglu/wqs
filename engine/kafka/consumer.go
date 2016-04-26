/*
Copyright 2009-2016 Weibo, Inc.

All files licensed under the Apache License, Version 2.0 (the "License");
you may not use these files except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package kafka

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"time"

	sarama "github.com/bsm/sarama-cluster"
	log "github.com/cihub/seelog"
)

type Consumer struct {
	topic    string
	group    string
	consumer *sarama.Consumer
}

const (
	timeout = 20 * time.Millisecond // 20ms超时
)

func NewConsumer(addrs []string, topic, group string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	consumer, err := sarama.NewConsumer(addrs, group, []string{topic}, config)
	if err != nil {
		log.Errorf("kafka consumer init failed, addrs:%s, err:%v", addrs, err)
		return nil, err
	}
	go func() {
		e := <-consumer.Errors()
		log.Warnf("kafka consumer err:%v", e)
	}()
	return &Consumer{topic, group, consumer}, nil
}

func (c *Consumer) Recv() ([]byte, error) {
	var data []byte
	var err error
	select {
	case msg := <-c.consumer.Messages():
		data = msg.Value
		c.consumer.MarkOffset(msg, "") // metedata的用处？
		dig := md5.Sum(msg.Value)
		log.Infof("R %s@%d@%d %s", msg.Topic, msg.Partition, msg.Offset, hex.EncodeToString(dig[:]))
	case <-time.After(timeout):
		err = errors.New("time out")
	}

	return data, err
}