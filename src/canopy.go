package main

import (
    "fmt"
    "io"
    "net/http"
    
    "code.google.com/p/go.net/websocket"
)

/*
type SingleConnection struct {
    conn *gosql.Conn
    cfg *gosql.ClusterConfig
}

func NewSingleConnection(cfg *gosql.ClusterConfig) gosql.ConnectionPool {
    addr := strings.TrimSpace(cfg.Hosts[0])
    if strings.Index(addr, ":") < 0 {
        addr = fmt.Sprintf("%s:%d", addr, cfg.DefaultPort)
    }

    connCfg := gosql.ConnConfig {
        ProtoVersion:   cfg.ProtoVersion,
        CQLVersion:     cfg.CQLVersion,
        Timeout:        cfg.Timeout,
        NumStreams:     cfg.NumStreams,
        Compressor:     cfg.Compressor,
        Authenticator:  cfg.Authenticator,
        Keepalive:      cfg.SocketKeepalive,
    }

    pool := SingleConnection{cfg:cfg}
    pool.conn = Connect(addr, connCfg, pool)
    return &pool
}

func (s *SingleConnection) HandleError(conn *gosql.Conn, err error, closed bool) {
    if closed {
        connCfg := gosql.ConnConfig{
            ProtoVersion:   cfg.ProtoVersion,
            CQLVersion:     cfg.CQLVersion,
            Timeout:        cfg.Timeout,
            NumStreams:     cfg.NumStreams,
            Compressor:     cfg.Compressor,
            Authenticator:  cfg.Authenticator,
            Keepalive:      cfg.SocketKeepalive,
        }
        s.conn = Connect(conn.Address(), connCfg, s)
    }
}

func (s *SingleConnection) Pick(qry *Query) *Conn {
    if s.conn.isClosed {
        return nil
    }
    return s.conn
}

func (s *SingleConnection) Size() int {
    return 1
}

func (s *SingleConnection) Close() {
    s.conn.Close()
}

func foo() {
    cluster := NewCluster("127.0.0.1")
    cluster.ConnPoolType = NewSingleConnection
    session, err := cluster.CreateSession()
}*/

// Echo the data received on the WebSocket
func EchoServer(ws *websocket.Conn) {
    io.Copy(ws, ws)
}


func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
    http.Handle("/echo", websocket.Handler(EchoServer))
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
