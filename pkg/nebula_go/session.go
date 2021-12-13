/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package nebula_go

import (
	"fmt"
	"sync"

	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
	"github.com/harrischu/nebula-wrapper/pkg/nebula_go/driver"
)

type timezoneInfo struct {
	offset int32
	name   []byte
}

type Session struct {
	sessionID  int64
	connection driver.Connection
	connPool   *ConnectionPool
	log        Logger
	mu         sync.Mutex
	timezoneInfo
}

// Execute returns the result of the given query as a ResultSet
func (s *Session) Execute(stmt string) (driver.ResultSet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.connection == nil {
		return nil, fmt.Errorf("failed to execute: Session has been released")
	}
	resp, err := s.connection.Execute(s.sessionID, stmt)
	if err == nil {
		return resp, nil
	}
	// Reconnect only if the tranport is closed
	err2, ok := err.(thrift.TransportException)
	if !ok {
		return nil, err
	}
	if err2.TypeID() == thrift.END_OF_FILE {
		_err := s.reConnect()
		if _err != nil {
			s.log.Error(fmt.Sprintf("Failed to reconnect, %s", _err.Error()))
			return nil, _err
		}
		s.log.Info(fmt.Sprintf("Successfully reconnect to host: "))

		// Execute with the new connetion
		resp, err := s.connection.Execute(s.sessionID, stmt)
		if err != nil {
			return nil, err
		}

		return resp, nil
	} else { // No need to reconnect
		s.log.Error(fmt.Sprintf("Error info: %s", err2.Error()))
		return nil, err2
	}
}

// ExecuteJson returns the result of the given query as a json string
// Date and Datetime will be returned in UTC
//	JSON struct:
// {
//     "results":[
//         {
//             "columns":[
//             ],
//             "data":[
//                 {
//                     "row":[
//                         "row-data"
//                     ],
//                     "meta":[
//                         "metadata"
//                     ]
//                 }
//             ],
//             "latencyInUs":0,
//             "spaceName":"",
//             "planDesc ":{
//                 "planNodeDescs":[
//                     {
//                         "name":"",
//                         "id":0,
//                         "outputVar":"",
//                         "description":{
//                             "key":""
//                         },
//                         "profiles":[
//                             {
//                                 "rows":1,
//                                 "execDurationInUs":0,
//                                 "totalDurationInUs":0,
//                                 "otherStats":{}
//                             }
//                         ],
//                         "branchInfo":{
//                             "isDoBranch":false,
//                             "conditionNodeId":-1
//                         },
//                         "dependencies":[]
//                     }
//                 ],
//                 "nodeIndexMap":{},
//                 "format":"",
//                 "optimize_time_in_us":0
//             },
//             "comment ":""
//         }
//     ],
//     "errors":[
//         {
//       		"code": 0,
//       		"message": ""
//         }
//     ]
// }
func (session *Session) ExecuteJson(stmt string) ([]byte, error) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.connection == nil {
		return nil, fmt.Errorf("failed to execute: Session has been released")
	}
	resp, err := session.connection.ExecuteJson(session.sessionID, stmt)
	if err == nil {
		return resp, nil
	}
	// Reconnect only if the tranport is closed
	err2, ok := err.(thrift.TransportException)
	if !ok {
		return nil, err
	}
	if err2.TypeID() == thrift.END_OF_FILE {
		_err := session.reConnect()
		if _err != nil {
			session.log.Error(fmt.Sprintf("Failed to reconnect, %s", _err.Error()))
			return nil, _err
		}
		session.log.Info(fmt.Sprintf("Successfully reconnect to host"))
		// Execute with the new connetion
		resp, err := session.connection.ExecuteJson(session.sessionID, stmt)
		if err != nil {
			return nil, err
		}
		return resp, nil
	} else { // No need to reconnect
		session.log.Error(fmt.Sprintf("Error info: %s", err2.Error()))
		return nil, err2
	}
}

func (session *Session) reConnect() error {
	newconnection, err := session.connPool.getIdleConn()
	if err != nil {
		err = fmt.Errorf(err.Error())
		return err
	}

	// Release connection to pool
	session.connPool.release(session.connection)
	session.connection = newconnection
	return nil
}

// Release logs out and releases connetion hold by session.
// The connection will be added into the activeConnectionQueue of the connection pool
// so that it could be reused.
func (session *Session) Release() {
	if session == nil {
		return
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.connection == nil {
		session.log.Warn("Session has been released")
		return
	}
	if err := session.connection.Signout(session.sessionID); err != nil {
		session.log.Warn(fmt.Sprintf("Sign out failed, %s", err.Error()))
	}
	// Release connection to pool
	session.connPool.release(session.connection)
	session.connection = nil
}
