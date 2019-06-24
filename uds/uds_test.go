package uds

import "fmt"
import "testing"
import "time"
import "os"

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestClient(t *testing.T) {
	conn := NewQueueConnection("test", 4095)
	client := NewClient(conn, 0.2)

	go func() {
		fmt.Println("wait for request")
		conn.other()
		fmt.Println("we have the request")
		conn.fromuser.PushBack([]byte{0x76, 0x22, 0x89, 0xab, 0xcd, 0xef})
	}()

	go func() {
		fmt.Println("enter")
		response := client.transfer_data(0x22, []byte{0x12, 0x34, 0x56})
		fmt.Println("we have response")
		eq(t, response.service_data["sequence_number_echo"], 0x22, "TestClient")
	}()

	time.Sleep(200)
}
