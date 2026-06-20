package collect

import "testing"

func TestTagRisk(t *testing.T) {
	tests := []struct {
		path     string
		expected RiskTag
	}{
		{"internal/auth/handler.go", RiskAuth},
		{"pkg/login/service.go", RiskAuth},
		{"middleware/session.go", RiskAuth},
		{"security/policy.go", RiskAuth},

		{"api/v2/users.go", RiskAPI},
		{"handler/create.go", RiskAPI},
		{"internal/middleware/rate_limit.go", RiskAPI},
		{"graphql/schema.go", RiskAPI},

		{"db/migration/001_init.sql", RiskData},
		{"internal/model/user.go", RiskData},
		{"repository/user_repo.go", RiskData},

		{"deploy/terraform/main.tf", RiskInfra},
		{".github/workflows/deploy.yml", RiskInfra},
		{"helm/values.yaml", RiskInfra},

		{"internal/utils/strings.go", RiskOther},
		{"cmd/main.go", RiskOther},
		{"README.md", RiskOther},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := TagRisk(tt.path)
			if got != tt.expected {
				t.Errorf("TagRisk(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestRiskOrder(t *testing.T) {
	if RiskOrder(RiskAuth) >= RiskOrder(RiskAPI) {
		t.Error("auth should have higher priority (lower number) than api")
	}
	if RiskOrder(RiskAPI) >= RiskOrder(RiskData) {
		t.Error("api should have higher priority than data")
	}
	if RiskOrder(RiskData) >= RiskOrder(RiskInfra) {
		t.Error("data should have higher priority than infra")
	}
	if RiskOrder(RiskInfra) >= RiskOrder(RiskOther) {
		t.Error("infra should have higher priority than other")
	}
}
