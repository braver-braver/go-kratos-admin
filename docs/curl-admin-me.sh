#!/usr/bin/env bash
set -euo pipefail

# Simple login + /admin/v1/me smoke test for troubleshooting 401/403.
# Defaults are for local dev; override via env vars as needed.
HOST="${HOST:-http://localhost:7788}"
USERNAME="${USERNAME:-admin}"
PASSWORD="${PASSWORD:-admin}"
CLIENT_ID="${CLIENT_ID:-admin-web}"

login_payload=$(cat <<EOF
{
  "grant_type": "password",
  "client_id": "${CLIENT_ID}",
  "username": "${USERNAME}",
  "password": "${PASSWORD}"
}
EOF
)

echo "## Login -> ${HOST}/admin/v1/login"
login_resp=$(curl -sS -w "\n%{http_code}\n" -X POST \
  -H "Content-Type: application/json" \
  -d "${login_payload}" \
  "${HOST}/admin/v1/login")

login_body=$(echo "${login_resp}" | head -n -1)
login_status=$(echo "${login_resp}" | tail -n 1)

echo "Status: ${login_status}"
echo "Body  : ${login_body}"

if [[ "${login_status}" != "200" ]]; then
  echo "Login failed; aborting." >&2
  exit 1
fi

access_token=$(echo "${login_body}" | jq -r '.accessToken')
refresh_token=$(echo "${login_body}" | jq -r '.refreshToken')

if [[ -z "${access_token}" || "${access_token}" == "null" ]]; then
  echo "No accessToken in login response; aborting." >&2
  exit 1
fi

echo
echo "## GET /admin/v1/me (Authorization: Bearer <access_token>)"
me_resp=$(curl -sS -w "\n%{http_code}\n" \
  -H "Authorization: Bearer ${access_token}" \
  "${HOST}/admin/v1/me")

me_body=$(echo "${me_resp}" | head -n -1)
me_status=$(echo "${me_resp}" | tail -n 1)

echo "Status: ${me_status}"
echo "Body  : ${me_body}"

echo
echo "access_token : ${access_token}"
echo "refresh_token: ${refresh_token}"
