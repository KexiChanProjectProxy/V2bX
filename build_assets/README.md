# V2bX Pre-built Binaries

Version: **0.4.1-sniff**

## About This Build

This build includes sniffing and domain strategy configuration options for all sing-box inbound protocols.

### New Features
- Traffic sniffing enabled by default
- Sniff override destination enabled
- Configurable sniff timeout (default: 300ms)
- Domain resolution strategy (default: prefer_ipv4)

## Available Binaries

### Linux AMD64 (x86_64)
- **File**: `V2bX-linux-amd64.tar.gz`
- **Architecture**: x86_64 (64-bit Intel/AMD)
- **Size**: ~29MB (compressed), 89MB (extracted)
- **Compatibility**: Ubuntu 22.04+, Debian 11+, CentOS 8+, and other modern Linux distributions

### Linux ARM64 (AArch64)
- **File**: `V2bX-linux-arm64.tar.gz`
- **Architecture**: ARM64/AArch64 (64-bit ARM)
- **Size**: ~27MB (compressed), 83MB (extracted)
- **Compatibility**: Ubuntu 22.04+ ARM64, Raspberry Pi 4+ (64-bit), AWS Graviton, and other ARM64 systems

## Installation

### 1. Download and Extract

**For AMD64:**
```bash
wget https://github.com/KexiChanProjectProxy/V2bX/raw/master/build_assets/V2bX-linux-amd64.tar.gz
tar -xzf V2bX-linux-amd64.tar.gz
chmod +x V2bX-linux-amd64
sudo mv V2bX-linux-amd64 /usr/local/bin/V2bX
```

**For ARM64:**
```bash
wget https://github.com/KexiChanProjectProxy/V2bX/raw/master/build_assets/V2bX-linux-arm64.tar.gz
tar -xzf V2bX-linux-arm64.tar.gz
chmod +x V2bX-linux-arm64
sudo mv V2bX-linux-arm64 /usr/local/bin/V2bX
```

### 2. Verify Installation

```bash
V2bX version
```

Expected output:
```
V2bX 0.4.1-sniff (A V2board backend based on multi core)
```

### 3. Verify Checksum (Optional but Recommended)

```bash
sha256sum -c SHA256SUMS
```

## Configuration

The new sniff options can be configured in your config file under `SingOptions`:

```json
{
  "Options": {
    "Core": "sing",
    "EnableSniff": true,
    "SniffOverrideDestination": true,
    "SniffTimeout": "300ms",
    "DomainStrategy": "prefer_ipv4"
  }
}
```

### Configuration Options

- **EnableSniff**: Enable traffic sniffing (default: `true`)
- **SniffOverrideDestination**: Override destination based on sniffed domain (default: `true`)
- **SniffTimeout**: Timeout for sniffing operation (default: `"300ms"`)
- **DomainStrategy**: Domain resolution strategy
  - `"prefer_ipv4"` - Prefer IPv4 addresses (default)
  - `"prefer_ipv6"` - Prefer IPv6 addresses
  - `"ipv4_only"` - Use IPv4 only
  - `"ipv6_only"` - Use IPv6 only

## Supported Protocols

These sniff options apply to all sing-box protocols:
- vmess
- vless
- shadowsocks
- trojan
- tuic
- anytls
- hysteria
- hysteria2

## Build Information

- **Build Date**: 2025-12-18
- **Go Version**: 1.25.0
- **CGO**: Disabled (statically linked)
- **Build Tags**: `sing xray hysteria2 with_quic with_grpc with_utls with_wireguard with_acme with_gvisor`
- **LDFLAGS**: `-s -w` (stripped and optimized)

## System Requirements

- **OS**: Linux (Ubuntu 22.04+, Debian 11+, CentOS 8+, or equivalent)
- **Kernel**: 3.10+ (recommended 4.4+)
- **RAM**: Minimum 512MB, recommended 1GB+
- **Disk**: 200MB free space

## SHA256 Checksums

```
784f300fa7abdef75b2399df4cb1c96eabb019dafcb964ed43b81a1ac225b0d2  V2bX-linux-amd64.tar.gz
9f570659d87e190b047aaf401ab9c5e3462239bc1f7a7737351fae5ad2f495c4  V2bX-linux-arm64.tar.gz
```

## Troubleshooting

### Permission Denied
If you get "permission denied" error, make sure the binary is executable:
```bash
chmod +x V2bX-linux-amd64  # or V2bX-linux-arm64
```

### Library Errors
These binaries are statically linked and should work without additional libraries. If you encounter issues, ensure you're running on a supported Linux distribution.

### Architecture Mismatch
Make sure you download the correct binary for your architecture:
- Use `uname -m` to check your architecture
  - `x86_64` → use amd64 binary
  - `aarch64` or `arm64` → use arm64 binary

## Support

- **Issues**: https://github.com/KexiChanProjectProxy/V2bX/issues
- **Original Project**: https://github.com/wyx2685/V2bX

## License

This project inherits the license from the original V2bX project.
