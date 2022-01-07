package ipc

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var mq *MsqQ

func MqInit(t *testing.T) {
	var err error
	os.Create("/tmp/mq")
	mq, err = Queue("/tmp/mq", 10,0)
	if err != nil {
		t.Errorf("err: %v", err)
		os.Exit(1)
	}
}

func TestMsqQ_Ctrl(t *testing.T) {
	MqInit(t)
	var buf msqid_ds
	err := mq.Ctrl(IPC_STAT, &buf)
	if err != nil {
		t.Errorf("ctrl failed, err: %v", err)
		return
	}

	t.Logf("%+v", buf)
}

func TestQueueSend(t *testing.T) {
	MqInit(t)

	for i := 0; i < 100; i++ {
		msg := fmt.Sprintf("idx: %v, send data: %v", i, time.Now())
		var data [256]uint8
		copy(data[:], msg)
		t.Logf("send: %v", msg)
		err := mq.Send(100, data)
		if err != nil {
			t.Errorf("send [%v] failed, err: %v", i, err)
		}
		time.Sleep(1 * time.Second)
	}
}

func TestQueueRecv(t *testing.T) {
	MqInit(t)
	for {
		data, err := mq.Recv(100)
		if err != nil {
			t.Errorf("recv failed, err: %v", err)
		} else {
			t.Logf("recv: %v", string(data))
		}
		time.Sleep(1 * time.Second)
	}
}
