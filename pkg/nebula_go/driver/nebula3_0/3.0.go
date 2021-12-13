package nebula3_0

import (
	"fmt"

	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
	"github.com/harrischu/nebula-wrapper/pkg/nebula_go/driver"
	"github.com/harrischu/nebula-wrapper/pkg/thrift/3.0/nebula"
	"github.com/harrischu/nebula-wrapper/pkg/thrift/3.0/nebula/graph"
)

type Driver struct{}

var _ driver.Driver = &Driver{}

type Connection struct {
	driver.DefaultConnection
}

type ResultSet struct {
	driver.DefaultResultSet
	resp     *graph.ExecutionResponse
	authResp *graph.AuthResponse
}

func (d *Driver) GetResultSet(v interface{}) driver.ResultSet {
	if _, ok := v.(*graph.ExecutionResponse); !ok {
		panic("cannot convert interface to response")
	}
	rs := &ResultSet{}
	rs.resp = v.(*graph.ExecutionResponse)
	return rs
}

func (d *Driver) NewConnection(host string, port int) driver.Connection {
	c := &Connection{}
	c.Host = host
	c.Port = port
	c.NebulaDriver = d
	c.Factory = func(t thrift.Transport, f thrift.ProtocolFactory) driver.GraphClient {
		client := graph.NewGraphServiceClientFactory(t, f)
		return client
	}
	return c
}

func (c *Connection) Execute(sessionID int64, stmt string) (driver.ResultSet, error) {
	resp, err := c.Graph.(*graph.GraphServiceClient).Execute(sessionID, []byte(stmt))
	if err == nil {
		return c.NebulaDriver.GetResultSet(resp), nil
	}
	// reopen the connection if timeout
	if _, ok := err.(thrift.TransportException); !ok {
		return nil, err
	}
	if err.(thrift.TransportException).TypeID() != thrift.TIMED_OUT {
		return nil, err
	}
	if err := c.Reopen(); err != nil {
		return nil, err
	}
	resp, err = c.Graph.(*graph.GraphServiceClient).Execute(sessionID, []byte(stmt))
	if err != nil {
		return nil, err
	}
	return c.NebulaDriver.GetResultSet(resp), nil
}

// Check connection to host address
func (c *Connection) ping() bool {
	_, err := c.Execute(0, "YIELD 1")
	return err == nil
}

func (c *Connection) ExecuteJson(sessionID int64, stmt string) ([]byte, error) {
	jsonResp, err := c.Graph.(*graph.GraphServiceClient).ExecuteJson(sessionID, []byte(stmt))
	if err == nil {
		return jsonResp, nil
	}
	if _, ok := err.(thrift.TransportException); !ok {
		return nil, err
	}
	// reopen the connection if timeout
	if err.(thrift.TransportException).TypeID() != thrift.TIMED_OUT {
		return nil, err
	}
	reopenErr := c.Reopen()
	if reopenErr != nil {
		return nil, reopenErr
	}
	return c.Graph.(*graph.GraphServiceClient).ExecuteJson(sessionID, []byte(stmt))
}

func (c *Connection) Authenticate(username, password string) (driver.AuthResultSet, error) {
	resp, err := c.Graph.(*graph.GraphServiceClient).Authenticate([]byte(username), []byte(password))
	if err != nil {
		err = fmt.Errorf("authentication fails, %s", err.Error())
		if e := c.Graph.Close(); e != nil {
			err = fmt.Errorf("fail to close transport, error: %s", e.Error())
		}
		return nil, err
	}
	if resp.ErrorCode != nebula.ErrorCode_SUCCEEDED {
		return nil, fmt.Errorf("fail to authenticate, error: %s", resp.ErrorMsg)
	}
	return resp, err
}

func (rs *ResultSet) GetErrorCode() driver.ErrorCode {
	e := rs.resp.GetErrorCode()
	return driver.ErrorCode(e)
}

func (rs *ResultSet) GetLatency() int32 {
	l := rs.resp.GetLatencyInUs()
	return int32(l)
}

func (rs *ResultSet) GetSpaceName() string {
	return string(rs.resp.GetSpaceName())
}

func (rs *ResultSet) GetErrorMsg() string {
	return string(rs.resp.GetErrorMsg())
}

func (rs *ResultSet) IsEmpty() bool {
	if !rs.resp.IsSetData() || len(rs.resp.Data.Rows) == 0 {
		return true
	}
	return false
}

func (rs *ResultSet) IsSucceed() bool {
	return rs.resp.GetErrorCode() == nebula.ErrorCode_SUCCEEDED

}
func (rs *ResultSet) IsPartialSucceed() bool {
	return rs.resp.GetErrorCode() == nebula.ErrorCode_E_PARTIAL_SUCCEEDED

}
func (rs *ResultSet) GetSessionID() int64 {
	return rs.authResp.GetSessionID()

}

// Returns the number of total rows
func (rs *ResultSet) GetRowSize() int {
	if rs.resp.Data == nil {
		return 0
	}
	return len(rs.resp.Data.Rows)
}

func (rs *ResultSet) IsSetData() bool {
	return rs.resp.Data != nil
}

func (rs *ResultSet) GetComment() string {
	if rs.resp.Comment == nil {
		return ""
	}
	return string(rs.resp.Comment)
}

func (res ResultSet) IsSetComment() bool {
	return res.resp.Comment != nil
}

func (res ResultSet) IsSetPlanDesc() bool {
	return res.resp.PlanDesc != nil
}

func (res ResultSet) GetPlanDesc() *graph.PlanDescription {
	return res.resp.PlanDesc
}