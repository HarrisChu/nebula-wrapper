package driver

import (
	"crypto/tls"
	"fmt"
	"math"
	"time"

	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
)

type GraphClient interface {
	Open() error
	IsOpen() bool
	Close() error
	Signout(sessionId int64) (err error)
}

type DefaultConnection struct {
	Host         string
	Port         int
	timeout      time.Duration
	ReturnedAt   time.Time // the connection was created or returned.
	NebulaDriver Driver
	Graph        GraphClient
	Factory      GraphFactory
	sslConfig    *tls.Config
}

type timezoneInfo struct {
	offset int32
	name   []byte
}

type GraphFactory func(t thrift.Transport, f thrift.ProtocolFactory) GraphClient

func (c *DefaultConnection) Open(timeout time.Duration, sslConfig *tls.Config) error {
	ip := c.Host
	port := c.Port
	newAdd := fmt.Sprintf("%s:%d", ip, port)
	c.timeout = timeout
	c.sslConfig = sslConfig
	bufferSize := 128 << 10
	frameMaxLength := uint32(math.MaxUint32)

	var err error
	var sock thrift.Transport
	if sslConfig != nil {
		sock, err = thrift.NewSSLSocketTimeout(newAdd, sslConfig, timeout)
	} else {
		sock, err = thrift.NewSocket(thrift.SocketAddr(newAdd), thrift.SocketTimeout(timeout))
	}
	if err != nil {
		return fmt.Errorf("failed to create a net.Conn-backed Transport,: %s", err.Error())
	}

	// Set transport buffer
	bufferedTranFactory := thrift.NewBufferedTransportFactory(bufferSize)
	transport := thrift.NewFramedTransportMaxLength(bufferedTranFactory.GetTransport(sock), frameMaxLength)
	pf := thrift.NewBinaryProtocolFactoryDefault()
	c.Graph = c.Factory(transport, pf)
	if err = c.Graph.Open(); err != nil {
		return fmt.Errorf("failed to open transport, error: %s", err.Error())
	}
	if !c.Graph.IsOpen() {
		return fmt.Errorf("transport is off")
	}
	return nil
}

func (c *DefaultConnection) Reopen() error {
	err := c.Close()
	if err != nil {
		return err
	}
	err = c.Open(c.timeout, c.sslConfig)
	return err
}

func (c *DefaultConnection) Close() error {
	return c.Graph.Close()

}
func (c *DefaultConnection) Release() {
	c.ReturnedAt = time.Now()
}
func (c *DefaultConnection) Ping() bool {
	return true

}
func (c *DefaultConnection) GetReturnedAt() time.Time {
	return c.ReturnedAt
}

func (c *DefaultConnection) Signout(sessionID int64) error {
	return nil
}

type DefaultResultSet struct {
	ColumnNames     []string
	ColNameIndexMap map[string]int
}

func (rs *DefaultResultSet) GetColNames() []string {
	return rs.ColumnNames
}
