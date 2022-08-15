package consistenthashing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	conf "github.com/amazingchow/consistent-hashing-service-provider/internal/config"
)

func TestExecutor(t *testing.T) {
	// 1. 初始化服务器列表
	fakeServers := []string{
		"192.168.1.10",
		"192.168.1.17",
		"192.168.1.48",
		"192.168.1.71",
		"192.168.1.92",
		"192.168.1.110",
		"192.168.1.111",
		"192.168.1.145",
	}

	fakeConf := &conf.ConsistentHashing{}
	fakeConf.VirReplicas = 20
	executor := NewExecutor("w1", fakeConf)
	executor.Start()

	// 2. 将服务器添加到哈希环中
	for _, srv := range fakeServers {
		err := executor.Join(srv)
		assert.Empty(t, err)
	}
	servers, err := executor.List()
	assert.Empty(t, err)
	assert.Equal(t, len(fakeServers), len(servers), fmt.Sprintf("期望添加%d台服务器，实际只添加了%d服务器", len(fakeServers), len(servers)))

	// 3. 执行100次键映射，检查是否都会定位到同一台服务器
	expectedServer, err := executor.Map("foo")
	assert.Empty(t, err)
	for i := 1; i < 100; i++ {
		server, _ := executor.Map("foo")
		assert.Equal(t, expectedServer, server, fmt.Sprintf("执行第%d次键映射时，期望定位到服务器%s，实际定位到服务器%s", i, expectedServer, server))
	}

	// 4. 再次检查
	expectedServer, err = executor.Map("bar")
	assert.Empty(t, err)
	for i := 1; i < 100; i++ {
		server, _ := executor.Map("bar")
		assert.Equal(t, expectedServer, server, fmt.Sprintf("执行第%d次键映射时，期望定位到服务器%s，实际定位到服务器%s", i, expectedServer, server))
	}

	// 5. 移除某台服务器，再次检查
	err = executor.Leave(expectedServer)
	assert.Empty(t, err)
	for i := 0; i < 100; i++ {
		server, _ := executor.Map("foo")
		assert.NotEqual(t, expectedServer, server, fmt.Sprintf("服务器%s已被移除，但是执行第%d次键映射时还是定位到了该服务器", expectedServer, i))
	}

	executor.Stop()
}
