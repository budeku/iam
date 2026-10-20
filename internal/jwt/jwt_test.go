package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestManager_SignAndParse(t *testing.T) {
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

	tokenStr, err := manager.SignAccessToken(context.Background(), userUUID, sessionID, name, time.Minute*5)
	if err != nil {
		t.Fatalf("签发 Token 失败: %v", err)
	}

	if tokenStr == "" {
		t.Fatal("生成的 Token 不能为空")
	}

	claims, err := manager.Parse(context.Background(), tokenStr)
	if err != nil {
		t.Fatalf("解析/验证 Token 失败: %v", err)
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
}
