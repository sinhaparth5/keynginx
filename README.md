# KeyNginx CLI - Phase 1

A simple SSL certificate generator CLI tool built with Go.

## Phase 1 Features

- ✅ Generate RSA private keys (2048, 3072, 4096 bits)
- ✅ Create self-signed SSL certificates
- ✅ Support for localhost and custom domains
- ✅ Configurable certificate validity period
- ✅ Proper file permissions (600 for private keys)
- ✅ Certificate validation and information display

## Installation

```bash
# Clone repository
git clone https://github.com/sinhaparth5/keynginx
cd keynginx

# Build
make build

# Or install directly
make install
```

## Usage

### Generate certificates for localhost
```bash
keynginx certs --domain localhost --out ./ssl
```

### Generate certificates for custom domain
```bash
keynginx certs --domain myapp.local --out ./ssl --key-size 4096 --validity 730
```

### Verbose output with certificate details
```bash
keynginx certs --domain example.com --out ./certs --verbose
```

### Full example with all options
```bash
keynginx certs \
  --domain myapp.local \
  --out ./ssl \
  --key-size 2048 \
  --validity 365 \
  --country US \
  --state CA \
  --city "San Francisco" \
  --organization "My Company" \
  --unit "IT Department" \
  --email admin@myapp.local \
  --overwrite \
  --verbose
```

## Development

```bash
# Install dependencies
make deps

# Format code
make fmt

# Run tests
make test

# Build
make build

# Test certificate generation
make test-certs
```

## Next Phases

- **Phase 2**: Nginx configuration generation
- **Phase 3**: Docker integration
- **Phase 4**: Advanced features and polish
```

This is **Phase 1 ONLY** - just the core certificate generation functionality with a clean CLI interface. Ready to build and test!