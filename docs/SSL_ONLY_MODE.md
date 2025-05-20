# EntityDB SSL-Only Mode

As of v2.11.1, EntityDB runs in SSL-only mode by default when using the `entitydbd.sh` daemon script.

## Configuration

The daemon script now:
- Enables SSL by default
- Uses certificates at:
  - Certificate: `/etc/ssl/certs/server.pem`
  - Private Key: `/etc/ssl/private/server.key`
- Runs HTTPS on port 8443
- Disables HTTP listening entirely

## Security Benefits

1. **No Unencrypted Traffic**: Eliminates risk of accidental HTTP connections
2. **Enforced Encryption**: All API calls and dashboard access require HTTPS
3. **Certificate Validation**: Server won't start without valid certificates

## Certificate Setup

If certificates don't exist, the server will prompt you to run:

```bash
sudo /opt/entitydb/share/tools/setup_ssl.sh
```

This will help you:
- Generate self-signed certificates (for testing)
- Configure existing certificates
- Set proper permissions

## Client Updates

Update all client configurations to use HTTPS:

```bash
# Old HTTP URLs
http://localhost:8085

# New HTTPS URLs
https://localhost:8443
```

For self-signed certificates, use `-k` with curl:

```bash
curl -k https://localhost:8443/api/v1/status
```

## Testing

Verify SSL-only mode:

```bash
/opt/entitydb/share/tests/test_ssl_only.sh
```

This confirms:
- HTTPS is accessible
- HTTP port is closed
- Authentication works over SSL

## Reverting to Dual Mode

To enable both HTTP and HTTPS, start the server directly:

```bash
./bin/entitydb --use-ssl --ssl-cert=/path/to/cert --ssl-key=/path/to/key
```

This will run:
- HTTPS on port 8443
- HTTP redirect on port 8085