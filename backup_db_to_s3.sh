#!/bin/bash
# Daily SQLite backup to S3. Runs ON THE SERVER (installed at
# /opt/link-shortener/backup_db_to_s3.sh, cron: daily 02:15 UTC as ubuntu).
#
# - Uses sqlite3 .backup for a consistent snapshot of the live DB (no service stop).
# - Uploads gzipped snapshot to s3://$BUCKET/daily/ (lifecycle expires these after 35 days).
# - On the 1st of the month, also copies to monthly/ (kept indefinitely, moves to
#   STANDARD_IA after 30 days).
set -euo pipefail

DB="/var/lib/link-shortener/database.db"
BUCKET="af-link-shortener-backups-111766607077-ap-south-1"
AWS="/usr/local/bin/aws"
STAMP=$(date -u +%Y%m%d)

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

SNAP="$TMP/database-$STAMP.db"
sqlite3 "$DB" ".backup '$SNAP'"
gzip "$SNAP"

"$AWS" s3 cp "$SNAP.gz" "s3://$BUCKET/daily/database-$STAMP.db.gz" --only-show-errors

if [ "$(date -u +%d)" = "01" ]; then
    "$AWS" s3 cp "$SNAP.gz" "s3://$BUCKET/monthly/database-$STAMP.db.gz" --only-show-errors
fi

echo "$(date -u +%FT%TZ) backup OK: database-$STAMP.db.gz ($(du -h "$SNAP.gz" | cut -f1))"
