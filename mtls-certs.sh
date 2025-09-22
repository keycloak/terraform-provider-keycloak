#!/usr/bin/env bash
set -euo pipefail

CMD="${1:-}"
TLS_DIR="${2:-testdata/tls}"

usage() {
  echo "Usage: $0 {create|clean} <certs-dir>"
  exit 1
}

require() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Error: required command '$1' not found in PATH" >&2
    exit 127
  }
}

create() {
  require openssl
  umask 077
  mkdir -p "${TLS_DIR}"

  echo ">>> Generating CA key and certificate"
  openssl genrsa -out "${TLS_DIR}/ca-key.pem" 4096
  openssl req -x509 -new -key "${TLS_DIR}/ca-key.pem" -sha256 -days 3650 \
    -subj "/C=US/ST=Unknown/L=Unknown/O=Unknown/OU=Unknown/CN=Dev Test Root" \
    -out "${TLS_DIR}/ca-cert.pem"

  echo ">>> Generating server key and certificate"
  openssl genrsa -out "${TLS_DIR}/server-key.pem" 2048
  openssl req -new -key "${TLS_DIR}/server-key.pem" \
    -subj "/C=US/ST=Unknown/L=Unknown/O=Unknown/OU=Unknown/CN=localhost" \
    -out "${TLS_DIR}/server.csr"
  cat > "${TLS_DIR}/server.ext" <<EOF
basicConstraints=CA:false
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:localhost,IP:127.0.0.1
EOF
  openssl x509 -req -in "${TLS_DIR}/server.csr" \
    -CA "${TLS_DIR}/ca-cert.pem" -CAkey "${TLS_DIR}/ca-key.pem" -CAserial "${TLS_DIR}/ca-cert.srl" -CAcreateserial \
    -out "${TLS_DIR}/server-cert.pem" -days 825 -sha256 -extfile "${TLS_DIR}/server.ext"

  echo ">>> Generating trusted client key and certificate"
  openssl genrsa -out "${TLS_DIR}/client-key.pem" 2048
  openssl req -new -key "${TLS_DIR}/client-key.pem" \
    -subj "/C=US/ST=Unknown/L=Unknown/O=Unknown/OU=Unknown/CN=trusted-client-mtls" \
    -out "${TLS_DIR}/client.csr"
  cat > "${TLS_DIR}/client.ext" <<EOF
basicConstraints=CA:false
keyUsage=critical,digitalSignature
extendedKeyUsage=clientAuth
subjectAltName=DNS:trusted-client-mtls
EOF
  openssl x509 -req -in "${TLS_DIR}/client.csr" \
    -CA "${TLS_DIR}/ca-cert.pem" -CAkey "${TLS_DIR}/ca-key.pem" -CAserial "${TLS_DIR}/ca-cert.srl" \
    -out "${TLS_DIR}/client-cert.pem" -days 825 -sha256 -extfile "${TLS_DIR}/client.ext"

  echo ">>> Generating untrusted client key and certificate"
  openssl genrsa -out "${TLS_DIR}/untrusted-client-key.pem" 2048
  openssl req -new -key "${TLS_DIR}/untrusted-client-key.pem" \
    -subj "/CN=untrusted-client-mtls" \
    -out "${TLS_DIR}/untrusted-client.csr"
  cat > "${TLS_DIR}/untrusted-client.ext" <<EOF
basicConstraints=CA:false
keyUsage=critical,digitalSignature
extendedKeyUsage=clientAuth
subjectAltName=DNS:untrusted-client-mtls
EOF
  # self-signed (not trusted by CA)
  openssl x509 -req -in "${TLS_DIR}/untrusted-client.csr" \
    -key "${TLS_DIR}/untrusted-client-key.pem" \
    -out "${TLS_DIR}/untrusted-client-cert.pem" -days 825 -sha256 -extfile "${TLS_DIR}/untrusted-client.ext"

  echo ">>> Cleaning up temporary files"
  rm -f \
    "${TLS_DIR}/server.csr" \
    "${TLS_DIR}/client.csr" \
    "${TLS_DIR}/untrusted-client.csr" \
    "${TLS_DIR}/server.ext" \
    "${TLS_DIR}/client.ext" \
    "${TLS_DIR}/untrusted-client.ext" \
    "${TLS_DIR}/ca-cert.srl" || true

  echo ">>> Done. Certificates are in ${TLS_DIR}"
}

clean() {
  umask 077
  echo ">>> Removing generated certs and keys in ${TLS_DIR}"
  rm -f \
    "${TLS_DIR}/ca-key.pem"     "${TLS_DIR}/ca-cert.pem" \
    "${TLS_DIR}/server-key.pem" "${TLS_DIR}/server-cert.pem" \
    "${TLS_DIR}/client-key.pem" "${TLS_DIR}/client-cert.pem" \
    "${TLS_DIR}/untrusted-client-key.pem" "${TLS_DIR}/untrusted-client-cert.pem" \
    "${TLS_DIR}/server.csr"     "${TLS_DIR}/client.csr" "${TLS_DIR}/untrusted-client.csr" \
    "${TLS_DIR}/server.ext"     "${TLS_DIR}/client.ext"  "${TLS_DIR}/untrusted-client.ext" \
    "${TLS_DIR}/ca-cert.srl" || true
  echo ">>> Done."
}

case "${CMD}" in
  create) create ;;
  clean)  clean ;;
  *) usage ;;
esac
