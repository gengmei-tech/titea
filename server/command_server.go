package server

import (
	"fmt"
	ns "github.com/gengmei-tech/titea/server/namespace"
	"github.com/gengmei-tech/titea/server/store"
	"strings"
)

// 对应redis的server类命令

// flush current db data 命名空间不删除
func flushdbCommand(c *Client) error {
	svr := store.InitServer(c.store)
	svr.FlushDB(c.environ)
	return c.writer.String("OK")
}

// 正式使用前的数据清理 命名空间不删除
func flushallCommand(c *Client) error {
	return c.writer.String("OK")
}

// 以client开头发命令 暂时 client list
func clientCommand(c *Client) error {
	sub := strings.ToLower(string(c.args[0]))
	if sub == "list" {
		return clientListCommand(c)
	}
	return c.writer.String("OK")
}

// list all client
func clientListCommand(c *Client) error {
	clients := c.server.getAllClients()
	resp := make([][]byte, len(clients))
	if len(clients) > 0 {
		for i, t := range clients {
			resp[i] = []byte(t)
		}
	}
	return c.writer.Array(resp)
}

// 返回系统信息 目前返回命名空间
// #service
// group.service: creater
func infoCommad(c *Client) error {
	var resp [][]byte
	resp = append(resp, []byte("# Server Info"))
	resp = append(resp, []byte(fmt.Sprintf("BuildTs: %s", store.BuildTs)))
	resp = append(resp, []byte(fmt.Sprintf("GitHash: %s", store.GitHash)))
	resp = append(resp, []byte(fmt.Sprintf("GitBranch: %s", store.GitBranch)))
	resp = append(resp, []byte(fmt.Sprintf("ReleaseVersion: %s", store.ReleaseVersion)))
	resp = append(resp, []byte(fmt.Sprintf("GoVersion: %s", store.GoVersion)))
	resp = append(resp, []byte("---------------------------"))
	resp = append(resp, []byte("# Group.Service | DBindex | Creator"))
	namespaces := ns.GetAllNamespace()
	for _, namespace := range namespaces {
		resp = append(resp, []byte(fmt.Sprintf("%s | %d | %s", namespace.Name, namespace.Index, namespace.Creator)))
	}
	resp = append(resp, []byte("---------------------------"))
	resp = append(resp, []byte("# Params"))
	resp = append(resp, c.server.getParams()...)
	resp = append(resp, []byte("---------------------------"))
	return c.writer.Array(resp)
}

// 只返回当前命名空间下数据数量
func dbsizeCommad(c *Client) error {
	svr := store.InitServer(c.store)
	total, err := svr.DBSize(c.environ)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(total))
}
