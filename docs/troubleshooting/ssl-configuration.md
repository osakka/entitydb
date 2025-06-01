# EntityDB SSL/TLS Configuration

EntityDB v2.11.0 introduces SSL/TLS support for secure communication between clients and the server.

## Quick Start

To enable SSL, start EntityDB with the `--use-ssl` flag:

```bash
./bin/entitydb --use-ssl --ssl-cert=/etc/ssl/certs/server.pem --ssl-key=/etc/ssl/private/server.key
```

## Configuration Options

| Flag | Default | Description |
|------|---------|-------------|
| `--use-ssl` | `false` | Enable SSL/TLS |
| `--ssl-port` | `8443` | HTTPS server port |
| `--ssl-cert` | `/etc/ssl/certs/server.pem` | Path to SSL certificate |
| `--ssl-key` | `/etc/ssl/private/server.key` | Path to SSL private key |
| `--port` | `8085` | HTTP server port (redirects to HTTPS when SSL is enabled) |

## SSL Modes

### 1. SSL-Only Mode (Recommended for Production)

When SSL is enabled, EntityDB runs:
- HTTPS server on the SSL port (default: 8443)
- HTTP redirect server on the standard port (default: 8085)

All HTTP requests are automatically redirected to HTTPS.

### 2. HTTP-Only Mode (Development)

When SSL is disabled (default), EntityDB runs:
- HTTP server only on the standard port (default: 8085)

## Setting Up SSL

### Option 1: Self-Signed Certificate (Testing)

Use the provided setup script:

```bash
sudo /opt/entitydb/share/tools/setup_ssl.sh
```

Select option 1 to generate a self-signed certificate.

### Option 2: Let's Encrypt Certificate (Production)

Install certbot and obtain a certificate:

```bash
sudo apt install certbot
sudo certbot certonly --standalone -d yourdomain.com
```

Then start EntityDB with the Let's Encrypt certificate:

```bash
./bin/entitydb --use-ssl \
  --ssl-cert=/etc/letsencrypt/live/yourdomain.com/fullchain.pem \
  --ssl-key=/etc/letsencrypt/live/yourdomain.com/privkey.pem
```

### Option 3: Commercial Certificate

1. Place your certificate and key files in secure locations
2. Set appropriate permissions:
   ```bash
   sudo chmod 600 /path/to/private.key
   sudo chmod 644 /path/to/certificate.pem
   ```
3. Start EntityDB with your certificate:
   ```bash
   ./bin/entitydb --use-ssl --ssl-cert=/path/to/certificate.pem --ssl-key=/path/to/private.key
   ```

## Client Configuration

### API Endpoints

When SSL is enabled, update your API endpoints:
- HTTP: `http://localhost:8085` → HTTPS: `https://localhost:8443`

### cURL Examples

```bash
# With self-signed certificate (skip verification)
curl -k https://localhost:8443/api/v1/status

# With valid certificate
curl https://yourdomain.com:8443/api/v1/status
```

### Python Example

```python
import requests

# For self-signed certificates
response = requests.get('https://localhost:8443/api/v1/status', verify=False)

# For valid certificates
response = requests.get('https://yourdomain.com:8443/api/v1/status')
```

## Testing SSL Configuration

Run the SSL test script:

```bash
/opt/entitydb/share/tests/test_ssl.sh
```

This script verifies:
1. HTTPS server is accessible
2. HTTP redirects to HTTPS
3. API authentication works over SSL
4. Certificate details

## Security Best Practices

1. **Use Valid Certificates**: Use certificates from trusted CAs for production
2. **Keep Keys Secure**: Ensure private keys have restrictive permissions (600)
3. **Regular Updates**: Keep certificates updated before expiration
4. **Strong Ciphers**: EntityDB uses Go's default TLS configuration with strong ciphers
5. **HSTS**: Consider adding HSTS headers for additional security

## Troubleshooting

### Common Issues

1. **Permission Denied**
   - Ensure the server has read access to certificate and key files
   - Check file ownership and permissions

2. **Certificate Errors**
   - Verify certificate and key match
   - Check certificate expiration date
   - Ensure certificate includes correct hostname

3. **Port Already in Use**
   - Check if another service is using the SSL port
   - Use a different port with `--ssl-port`

### Debug Mode

Enable debug logging to troubleshoot SSL issues:

```bash
./bin/entitydb --use-ssl --log-level=debug
```

## Performance Considerations

SSL/TLS adds minimal overhead:
- ~1-2ms latency per request
- ~5-10% CPU increase under load
- Negligible impact on throughput

For maximum performance, consider:
- Using HTTP/2 (automatically enabled with Go's TLS)
- Session resumption (handled by Go's TLS)
- Hardware acceleration (if available)

## Future Enhancements

Planned SSL features:
- Client certificate authentication
- Certificate hot-reloading
- ACME/Let's Encrypt integration
- Configurable TLS versions and ciphers