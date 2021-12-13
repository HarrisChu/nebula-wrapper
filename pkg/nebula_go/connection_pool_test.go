package nebula_go

import "testing"

var poolAddress = []HostAddress{
	{
		Host: "192.168.8.152",
		Port: 9669,
	},
	// {
	// 	Host: "127.0.0.1",
	// 	Port: 3700,
	// },
	// {
	// 	Host: "127.0.0.1",
	// 	Port: 3701,
	// },
}

var nebulaLog = DefaultLogger{}

// Create default configs
var testPoolConfig = GetDefaultConf()

func TestWrapper(t *testing.T) {
	pool, err := NewConnectionPool("3.0", poolAddress, testPoolConfig, nebulaLog)
	if err != nil {
		t.Fatal(err)
	}
	session, err := pool.GetSession("root", "nebula")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := session.Execute("show hosts")
	t.Log(resp.GetLatency())
	if err != nil {
		t.Fatal(err)
	}
	resp, err = session.Execute("show hosts")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.IsSucceed())
	t.Log(resp.GetLatency())
	t.Fatal(1)
}
