#!/usr/bin/env bash
set -euo pipefail

IP="${1:-192.168.1.111}"
DNS_NAME="${2:-}"
OUT_DIR="${3:-certs}"

mkdir -p "${OUT_DIR}"

OPENSSL_CFG="$(mktemp)"
trap 'rm -f "${OPENSSL_CFG}"' EXIT

{
  echo "[ req ]"
  echo "default_bits = 2048"
  echo "prompt = no"
  echo "default_md = sha256"
  echo "x509_extensions = v3_req"
  echo "distinguished_name = dn"
  echo
  echo "[ dn ]"
  echo "C = US"
  echo "ST = Local"
  echo "L = Local"
  echo "O = OCPP Local"
  echo "OU = Dev"
  echo "CN = ${IP}"
  echo
  echo "[ v3_req ]"
  echo "subjectAltName = @alt_names"
  echo
  echo "[ alt_names ]"
  echo "IP.1 = ${IP}"
  if [[ -n "${DNS_NAME}" ]]; then
    echo "DNS.1 = ${DNS_NAME}"
  fi
} > "${OPENSSL_CFG}"

openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
  -keyout "${OUT_DIR}/ocpp_wss.key" \
  -out "${OUT_DIR}/ocpp_wss.crt" \
  -config "${OPENSSL_CFG}"

echo "Generated:"
echo "  ${OUT_DIR}/ocpp_wss.crt"
echo "  ${OUT_DIR}/ocpp_wss.key"
echo
echo "Start backend with:"
echo "  OCPP_TLS_CERT=${OUT_DIR}/ocpp_wss.crt OCPP_TLS_KEY=${OUT_DIR}/ocpp_wss.key OCPP_TLS_PORT=9443 go run cmd/server/main.go"
echo
echo "Charger URL example:"
echo "  wss://${IP}:9443/my-test"
