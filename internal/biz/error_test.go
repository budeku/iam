package biz

import (
	"errors"
	"fmt"
	"testing"
)

func mockStoreQuery() error {
	// 模拟拼接后的错误
	return fmt.Errorf("%w: 测试 errors.Is", ErrNotFound)
}

func TestErrorWrappingAndIs(t *testing.T) {
	err := mockStoreQuery()

	expectedMsg := "信息不存在: 测试 errors.Is"
	if err.Error() != expectedMsg {
		t.Errorf("期望错误文本是:\n%s\n实际得到:\n%s", expectedMsg, err.Error())
	} else {
		t.Log("错误文本拼接成功:", err.Error())
	}

	// 使用 errors.Is 检查
	if errors.Is(err, ErrNotFound) {
		t.Log("errors.Is 验证成功：成功识别出了 biz.ErrNotFound！")
	} else {
		t.Error("errors.Is 识别失败")
	}
}
