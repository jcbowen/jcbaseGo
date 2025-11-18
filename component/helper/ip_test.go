package helper

import "testing"

func TestAliCdnRealIpPriority(t *testing.T) {
	headers := map[string]string{
		"Ali-Cdn-Real-Ip": "183.227.122.77",
		"X-Real-IP":       "221.195.216.20",
		"X-Forwarded-For": "183.227.122.77, 221.195.216.20",
	}
	ip := GetRealIPFromHeaders(headers)
	if ip != "183.227.122.77" {
		t.Fatalf("expected Ali-Cdn-Real-Ip, got %s", ip)
	}
}

func TestCFConnectingIPPriority(t *testing.T) {
	headers := map[string]string{
		"CF-Connecting-IP": "203.0.113.10",
		"X-Forwarded-For":  "198.51.100.20, 203.0.113.10",
	}
	ip := GetRealIPFromHeaders(headers)
	if ip != "203.0.113.10" {
		t.Fatalf("expected CF-Connecting-IP, got %s", ip)
	}
}

func TestXForwardedForFirstValid(t *testing.T) {
	headers := map[string]string{
		"X-Forwarded-For": "183.227.122.77, 221.195.216.20",
	}
	ip := GetRealIPFromHeaders(headers)
	if ip != "183.227.122.77" {
		t.Fatalf("expected first valid from XFF, got %s", ip)
	}
}

func TestFallbackToXRealIP(t *testing.T) {
	headers := map[string]string{
		"X-Real-IP": "198.51.100.2",
	}
	ip := GetRealIPFromHeaders(headers)
	if ip != "198.51.100.2" {
		t.Fatalf("expected X-Real-IP, got %s", ip)
	}
}

func TestIgnoreNonIPHeaders(t *testing.T) {
	headers := map[string]string{
		"X-Forwarded-Host": "api.example.com:443",
	}
	ip := GetRealIPFromHeaders(headers)
	if ip != "" {
		t.Fatalf("expected empty when only non-IP headers present, got %s", ip)
	}
}
