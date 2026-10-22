package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestManager_SignAndParseWithFingerprint(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成测试私钥失败: %v", err)
	}

	kid := uuid.NewString()
	provider := NewMemoryKeyProvider(kid, priv)

	manager := New(
		provider,
		WithIssuer("test-issuer"),
		WithAudience("test-audience"),
	)

	userUUID := uuid.New()
	sessionID := uuid.New()
	name := "张三"
	appID := 1

	// 模拟客户端的初始 IP 和 UA
	clientIP := "192.168.1.100"
	clientUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0"

	tokenStr, err := manager.SignAccessToken(context.Background(), userUUID, sessionID, name, appID, time.Minute*5, clientIP, clientUA)
	if err != nil {
		t.Fatalf("签发 Token 失败: %v", err)
	}

	if tokenStr == "" {
		t.Fatal("生成的 Token 不能为空")
	}

	claims, err := manager.Parse(context.Background(), tokenStr, clientIP, clientUA)
	if err != nil {
		t.Fatalf("使用相同环境解析/验证 Token 失败: %v", err)
	}

	if claims.UserUUID != userUUID {
		t.Errorf("期望 UserUUID 为 %v，实际为 %v", userUUID, claims.UserUUID)
	}
	if claims.SessionID != sessionID {
		t.Errorf("期望 SessionID 为 %v，实际为 %v", sessionID, claims.SessionID)
	}
	if claims.Name != name {
		t.Errorf("期望 Name 为 %s，实际为 %s", name, claims.Name)
	}

	switchedIP := "192.168.1.200"
	_, err = manager.Parse(context.Background(), tokenStr, switchedIP, clientUA)
	if err != nil {
		t.Errorf("期望 C 网段内 IP 变动可以通过校验，实际报错: %v", err)
	}

	attackerIP := "10.0.0.88"
	attackerUA := "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)"
	_, err = manager.Parse(context.Background(), tokenStr, attackerIP, attackerUA)
	if err == nil {
		t.Fatal("期望在环境指纹不匹配时拦截并报错，实际校验通过了")
	}
	if err != ErrFingerprintMismatch {
		t.Errorf("期望错误类型为 ErrFingerprintMismatch，实际为: %v", err)
	}
}
