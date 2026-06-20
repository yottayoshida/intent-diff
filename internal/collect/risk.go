package collect

import (
	"strings"
)

// TagRisk assigns a risk tag to a file based on its path.
func TagRisk(path string) RiskTag {
	lower := strings.ToLower(path)

	if matchesAny(lower, authPatterns) {
		return RiskAuth
	}
	if matchesAny(lower, apiPatterns) {
		return RiskAPI
	}
	if matchesAny(lower, dataPatterns) {
		return RiskData
	}
	if matchesAny(lower, infraPatterns) {
		return RiskInfra
	}
	return RiskOther
}

var authPatterns = []string{
	"auth", "login", "logout", "session", "token",
	"password", "credential", "oauth", "jwt", "saml",
	"permission", "rbac", "acl", "security",
}

var apiPatterns = []string{
	"api/", "handler", "endpoint", "route", "controller",
	"middleware", "graphql", "grpc", "proto",
	"openapi", "swagger",
}

var dataPatterns = []string{
	"migration", "schema", "model", "entity",
	"database", "db/", "/store", "repository",
	"/query", "sql",
}

var infraPatterns = []string{
	"deploy", "infra", "terraform", "helm",
	"docker", "k8s", "kubernetes", "ci/", "cd/",
	".github/workflows",
}

func matchesAny(path string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(path, p) {
			return true
		}
	}
	return false
}
