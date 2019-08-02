package wolk

import (
	"fmt"
	"testing"
)

func TestPolicy(t *testing.T) {
	policy := &Policy{}
	policy_file := DefaultPolicyWolkFile
	LoadPolicy(policy_file, policy)
	fmt.Printf("Policy: %v", policy)
}
