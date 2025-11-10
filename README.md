# NFPCF - NF Profile Cache Function

NFPCF is a caching proxy for the NRF (Network Repository Function) in 5G core networks. It caches NF profiles to reduce load on the backend NRF and improve discovery performance.

## Features

- **NF Registration**: Transparent pass-through to backend NRF with caching
- **NF Discovery**: Cache-first lookup with NRF fallback
- **NF Deregistration**: Cache invalidation with NRF pass-through
- **NF Update**: Cache invalidation with NRF pass-through
- **TTL-based Cache**: Automatic expiration of stale entries
- **Type Indexing**: Fast lookup by NF type

## Architecture

```
┌──────────────┐
│  NF Clients  │
│ (AMF/SMF/...) │
└──────┬───────┘
       │
       ▼
┌──────────────────┐
│     NFPCF        │
│  ┌────────────┐  │
│  │   Cache    │  │
│  └────────────┘  │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│       NRF        │
└──────────────────┘
```

## Build

```bash
make build
```

## Run

```bash
make run
```

Or manually:

```bash
./bin/nfpcf -c ./config/nfpcfcfg.yaml
```

## Configuration

Edit `config/nfpcfcfg.yaml`:

```yaml
server:
  bindAddr: 0.0.0.0:8000

nrf:
  url: http://nrf:8000

cache:
  ttl: 300000000000  # 5 minutes in nanoseconds

logger:
  level: info
```

## Docker

Build:
```bash
docker build -t nfpcf:latest .
```

Run:
```bash
docker run -p 8000:8000 \
  -e NRF_URL=http://nrf:8000 \
  nfpcf:latest
```

## API Endpoints

### NF Management

- `PUT /nnrf-nfm/v1/nf-instances/:nfInstanceID` - Register NF
- `GET /nnrf-nfm/v1/nf-instances/:nfInstanceID` - Get NF profile
- `DELETE /nnrf-nfm/v1/nf-instances/:nfInstanceID` - Deregister NF
- `PATCH /nnrf-nfm/v1/nf-instances/:nfInstanceID` - Update NF

### NF Discovery

- `GET /nnrf-disc/v1/nf-instances?target-nf-type=...` - Discover NFs

## Testing

Point your NF clients to NFPCF instead of NRF:

```yaml
# In AMF/SMF config
nrfUri: http://nfpcf:8000
```

## License

Same as free5GC project
