package conf

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestCoreConfigOnlySing verifies CoreConfig accepts "sing" type and populates SingConfig.
func TestCoreConfigOnlySing(t *testing.T) {
	raw := `{"Type":"sing","Log":{"Disable":true,"Level":"error","Timestamp":false}}`
	var c CoreConfig
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatalf("unmarshal sing CoreConfig: %v", err)
	}
	if c.Type != "sing" {
		t.Errorf("expected Type=sing, got %q", c.Type)
	}
	if c.SingConfig == nil {
		t.Fatal("SingConfig should be populated for type=sing")
	}
}

func TestCoreConfigUnknownTypeIsIgnored(t *testing.T) {
	// Unknown type (e.g. former "xray") should not error — SingConfig stays nil.
	raw := `{"Type":"xray"}`
	var c CoreConfig
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatalf("unexpected error for unknown type: %v", err)
	}
	if c.SingConfig != nil {
		t.Error("SingConfig should be nil for unknown type")
	}
}

func TestCoreConfigNoXrayHy2Fields(t *testing.T) {
	// Verify CoreConfig struct does not expose xray/hy2 fields.
	c := CoreConfig{Type: "sing"}
	b, _ := json.Marshal(c)
	s := string(b)
	for _, forbidden := range []string{"XrayConfig", "Hysteria2Config"} {
		if strings.Contains(s, forbidden) {
			t.Errorf("CoreConfig should not contain %q field, got: %s", forbidden, s)
		}
	}
}

// TestOptionsOnlySing verifies Options.UnmarshalJSON populates SingOptions for "sing" core.
func TestOptionsOnlySing(t *testing.T) {
	raw := `{"Core":"sing","ListenIP":"127.0.0.1","EnableTFO":true}`
	var o Options
	if err := json.Unmarshal([]byte(raw), &o); err != nil {
		t.Fatalf("unmarshal sing Options: %v", err)
	}
	if o.Core != "sing" {
		t.Errorf("expected Core=sing, got %q", o.Core)
	}
	if o.SingOptions == nil {
		t.Fatal("SingOptions should be set for Core=sing")
	}
}

func TestOptionsUnknownCoreCleared(t *testing.T) {
	// For unknown cores the Core field is reset to "" and RawOptions is populated.
	raw := `{"Core":"xray","SomeField":"val"}`
	var o Options
	if err := json.Unmarshal([]byte(raw), &o); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Core != "" {
		t.Errorf("unknown core should reset Core to \"\", got %q", o.Core)
	}
	if o.SingOptions != nil {
		t.Error("SingOptions should be nil for unknown core")
	}
	if len(o.RawOptions) == 0 {
		t.Error("RawOptions should hold original data for unknown core")
	}
}

func TestOptionsNoXrayHy2Fields(t *testing.T) {
	o := Options{}
	b, _ := json.Marshal(o)
	s := string(b)
	for _, forbidden := range []string{"XrayOptions", "Hysteria2ConfigPath", "HysteriaConfigPath"} {
		if strings.Contains(s, forbidden) {
			t.Errorf("Options should not contain %q field", forbidden)
		}
	}
}

func TestSingConfigDefaults(t *testing.T) {
	sc := NewSingConfig()
	if sc.LogConfig.Level != "error" {
		t.Errorf("expected default log level=error, got %q", sc.LogConfig.Level)
	}
	if sc.NtpConfig.Server != "time.apple.com" {
		t.Errorf("expected default NTP server=time.apple.com, got %q", sc.NtpConfig.Server)
	}
}

func TestSingOptionsDefaults(t *testing.T) {
	so := NewSingOptions()
	if !so.SniffEnabled {
		t.Error("expected SniffEnabled=true by default")
	}
	if !so.SniffOverrideDestination {
		t.Error("expected SniffOverrideDestination=true by default")
	}
	if so.FallBackConfigs == nil {
		t.Error("FallBackConfigs should not be nil by default")
	}
	if so.Multiplex == nil {
		t.Error("Multiplex should not be nil by default")
	}
}
