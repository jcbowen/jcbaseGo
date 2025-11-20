package helper_test

import (
    "testing"
    helper "github.com/jcbowen/jcbaseGo/component/helper"
    "github.com/jcbowen/jcbaseGo/component/debugger"
)

// TestCheckAndSetDefault_NamedIntTypes_External 包外测试命名整型（包含 debugger.LogLevel）
func TestCheckAndSetDefault_NamedIntTypes_External(t *testing.T) {
    type C2 struct {
        Level debugger.LogLevel `default:"2"`
    }
    c2 := &C2{}
    if err := helper.CheckAndSetDefault(c2); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if c2.Level != debugger.LogLevel(2) {
        t.Fatalf("expected LogLevel=2, got %v", c2.Level)
    }
}