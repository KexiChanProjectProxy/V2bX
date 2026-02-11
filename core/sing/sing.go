package sing

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/log"

	"github.com/InazumaV/V2bX/conf"
	vCore "github.com/InazumaV/V2bX/core"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
)

var _ vCore.Core = (*Sing)(nil)

type DNSConfig struct {
	Servers []map[string]interface{} `json:"servers"`
	Rules   []map[string]interface{} `json:"rules"`
}

type Sing struct {
	box                       *box.Box
	ctx                       context.Context
	hookServer                *HookServer
	router                    adapter.Router
	logFactory                log.Factory
	users                     *UserMap
	nodeReportMinTrafficBytes map[string]int64
	originalPath              string
	originalPathRefresh       int
	originalPathCancel        context.CancelFunc
}

type UserMap struct {
	uidMap  map[string]int
	mapLock sync.RWMutex
}

func init() {
	vCore.RegisterCore("sing", New)
}

// loadOriginalConfig loads config from either a URL or local file path
func loadOriginalConfig(path string) ([]byte, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch remote config: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("remote config returned status %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read remote config: %w", err)
		}
		return data, nil
	}

	return os.ReadFile(path)
}

func New(c *conf.CoreConfig) (vCore.Core, error) {
	ctx := context.Background()
	ctx = box.Context(ctx, include.InboundRegistry(), include.OutboundRegistry(), include.EndpointRegistry(), include.DNSTransportRegistry(), include.ServiceRegistry())
	options := option.Options{}
	if len(c.SingConfig.OriginalPath) != 0 {
		data, err := loadOriginalConfig(c.SingConfig.OriginalPath)
		if err != nil {
			return nil, fmt.Errorf("read original config error: %s", err)
		}
		options, err = json.UnmarshalExtendedContext[option.Options](ctx, data)
		if err != nil {
			return nil, fmt.Errorf("unmarshal original config error: %s", err)
		}
	}
	options.Log = &option.LogOptions{
		Disabled:  c.SingConfig.LogConfig.Disabled,
		Level:     c.SingConfig.LogConfig.Level,
		Timestamp: c.SingConfig.LogConfig.Timestamp,
		Output:    c.SingConfig.LogConfig.Output,
	}
	options.NTP = &option.NTPOptions{
		Enabled:       c.SingConfig.NtpConfig.Enable,
		WriteToSystem: true,
		ServerOptions: option.ServerOptions{
			Server:     c.SingConfig.NtpConfig.Server,
			ServerPort: c.SingConfig.NtpConfig.ServerPort,
		},
	}
	os.Setenv("SING_DNS_PATH", "")
	b, err := box.New(box.Options{
		Context: ctx,
		Options: options,
	})
	if err != nil {
		return nil, err
	}
	hs := &HookServer{
		counter: sync.Map{},
	}
	b.Router().AppendTracker(hs)
	return &Sing{
		ctx:        b.Router().GetCtx(),
		box:        b,
		hookServer: hs,
		router:     b.Router(),
		logFactory: b.LogFactory(),
		users: &UserMap{
			uidMap: make(map[string]int),
		},
		nodeReportMinTrafficBytes: make(map[string]int64),
		originalPath:              c.SingConfig.OriginalPath,
		originalPathRefresh:       c.SingConfig.OriginalPathRefresh,
	}, nil
}

func (b *Sing) Start() error {
	err := b.box.Start()
	if err != nil {
		return err
	}

	// Start refresh goroutine if configured
	if b.originalPathRefresh > 0 && len(b.originalPath) > 0 {
		if strings.HasPrefix(b.originalPath, "http://") || strings.HasPrefix(b.originalPath, "https://") {
			ctx, cancel := context.WithCancel(context.Background())
			b.originalPathCancel = cancel

			go func() {
				ticker := time.NewTicker(time.Duration(b.originalPathRefresh) * time.Second)
				defer ticker.Stop()

				var lastHash [32]byte
				// Compute initial hash
				if data, err := loadOriginalConfig(b.originalPath); err == nil {
					lastHash = sha256.Sum256(data)
				}

				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						data, err := loadOriginalConfig(b.originalPath)
						if err != nil {
							b.logFactory.NewLogger("config-refresh").Error("failed to fetch remote config: ", err)
							continue
						}

						currentHash := sha256.Sum256(data)
						if currentHash != lastHash {
							b.logFactory.NewLogger("config-refresh").Info("remote config changed, restart needed")
							lastHash = currentHash
						}
					}
				}
			}()
		}
	}

	return nil
}

func (b *Sing) Close() error {
	if b.originalPathCancel != nil {
		b.originalPathCancel()
	}
	return b.box.Close()
}

func (b *Sing) Protocols() []string {
	return []string{
		"vmess",
		"vless",
		"shadowsocks",
		"trojan",
		"tuic",
		"anytls",
		"hysteria",
		"hysteria2",
	}
}

func (b *Sing) Type() string {
	return "sing"
}
