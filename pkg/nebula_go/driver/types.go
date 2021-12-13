package driver

import (
	"crypto/tls"
	"time"
)

type ErrorCode int64

const (
	ErrorCode_SUCCEEDED ErrorCode = iota
	ErrorCode_E_DISCONNECTED
	ErrorCode_E_FAIL_TO_CONNECT
	ErrorCode_E_RPC_FAILURE
	ErrorCode_E_BAD_USERNAME_PASSWORD
	ErrorCode_E_SESSION_INVALID
	ErrorCode_E_SESSION_TIMEOUT
	ErrorCode_E_SYNTAX_ERROR
	ErrorCode_E_EXECUTION_ERROR
	ErrorCode_E_STATEMENT_EMPTY
	ErrorCode_E_USER_NOT_FOUND
	ErrorCode_E_BAD_PERMISSION
	ErrorCode_E_SEMANTIC_ERROR
)

type ConnectionMaker interface {
	NewConnection(host string, port int) Connection
}

type ResultSetMaker interface {
	GetResultSet(v interface{}) ResultSet
}

type Driver interface {
	ConnectionMaker
	ResultSetMaker
}

type Connection interface {
	Open(timeout time.Duration, sslConfig *tls.Config) error
	Close() error
	Release()
	Ping() bool
	GetReturnedAt() time.Time
	Execute(sessionID int64, stmt string) (ResultSet, error)
	ExecuteJson(sessionID int64, stmt string) ([]byte, error)
	Signout(sessionID int64) error
	Authenticate(username, password string) (AuthResultSet, error)
}

type AuthResultSet interface {
	GetSessionID() int64
	GetTimeZoneOffsetSeconds() int32
	GetTimeZoneName() []byte
}

type ResultSet interface {
	GetErrorCode() ErrorCode
	GetLatency() int32
	GetSpaceName() string
	GetErrorMsg() string
	IsEmpty() bool
	IsSucceed() bool
	IsPartialSucceed() bool
	GetRowSize() int
	GetColNames() []string
	IsSetData() bool
	IsSetComment() bool
	GetComment() string
}

type Rows interface {
	Len() int32
	Scan(v ...interface{}) error
}
