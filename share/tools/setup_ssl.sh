#!/bin/bash
# Setup SSL for EntityDB

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default paths
DEFAULT_CERT_PATH="/etc/ssl/certs/server.pem"
DEFAULT_KEY_PATH="/etc/ssl/private/server.key"
DEFAULT_SSL_DIR="/etc/ssl"

echo -e "${YELLOW}EntityDB SSL Setup Utility${NC}"
echo "=========================="
echo

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Error: This script must be run as root to create certificates${NC}"
    exit 1
fi

# Function to generate self-signed certificate
generate_self_signed() {
    local cert_path=$1
    local key_path=$2
    local hostname=$3
    
    echo -e "${YELLOW}Generating self-signed certificate...${NC}"
    
    # Create directories if they don't exist
    mkdir -p $(dirname "$cert_path")
    mkdir -p $(dirname "$key_path")
    
    # Generate private key
    openssl genrsa -out "$key_path" 2048
    
    # Generate certificate signing request
    openssl req -new -key "$key_path" -out /tmp/server.csr \
        -subj "/C=US/ST=State/L=City/O=Organization/OU=IT/CN=$hostname"
    
    # Generate self-signed certificate
    openssl x509 -req -days 365 -in /tmp/server.csr -signkey "$key_path" -out "$cert_path"
    
    # Set proper permissions
    chmod 600 "$key_path"
    chmod 644 "$cert_path"
    
    # Clean up
    rm -f /tmp/server.csr
    
    echo -e "${GREEN}Self-signed certificate generated successfully!${NC}"
}

# Main menu
echo "SSL Setup Options:"
echo "1. Generate self-signed certificate (for testing)"
echo "2. Use existing certificate"
echo "3. Display EntityDB SSL configuration flags"
echo
read -p "Select option (1-3): " option

case $option in
    1)
        echo
        read -p "Enter hostname (default: localhost): " hostname
        hostname=${hostname:-localhost}
        
        read -p "Certificate path (default: $DEFAULT_CERT_PATH): " cert_path
        cert_path=${cert_path:-$DEFAULT_CERT_PATH}
        
        read -p "Private key path (default: $DEFAULT_KEY_PATH): " key_path
        key_path=${key_path:-$DEFAULT_KEY_PATH}
        
        generate_self_signed "$cert_path" "$key_path" "$hostname"
        
        echo
        echo -e "${GREEN}Certificate generated!${NC}"
        echo "To start EntityDB with SSL, use:"
        echo "./bin/entitydb --use-ssl --ssl-cert=$cert_path --ssl-key=$key_path"
        ;;
        
    2)
        echo
        read -p "Enter certificate path: " cert_path
        read -p "Enter private key path: " key_path
        
        # Verify files exist
        if [ ! -f "$cert_path" ]; then
            echo -e "${RED}Error: Certificate file not found: $cert_path${NC}"
            exit 1
        fi
        
        if [ ! -f "$key_path" ]; then
            echo -e "${RED}Error: Private key file not found: $key_path${NC}"
            exit 1
        fi
        
        echo
        echo -e "${GREEN}Certificate files verified!${NC}"
        echo "To start EntityDB with SSL, use:"
        echo "./bin/entitydb --use-ssl --ssl-cert=$cert_path --ssl-key=$key_path"
        ;;
        
    3)
        echo
        echo "EntityDB SSL Configuration Flags:"
        echo "  --use-ssl                Enable SSL/TLS"
        echo "  --ssl-port=8443         SSL server port (default: 8443)"
        echo "  --ssl-cert=path         Path to SSL certificate file"
        echo "  --ssl-key=path          Path to SSL private key file"
        echo
        echo "Example:"
        echo "  ./bin/entitydb --use-ssl --ssl-cert=/etc/ssl/certs/server.pem --ssl-key=/etc/ssl/private/server.key"
        ;;
        
    *)
        echo -e "${RED}Invalid option${NC}"
        exit 1
        ;;
esac

echo
echo -e "${YELLOW}Note: Remember to update your clients to use https:// instead of http://${NC}"