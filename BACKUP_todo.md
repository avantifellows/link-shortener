# Database Backup Setup TODO

## Required Information to Gather

### AWS Configuration
- [ ] S3 bucket name for backups
- [ ] AWS region
- [ ] IAM credentials (access key/secret or IAM role if EC2 has one)

### Backup Strategy
- [ ] Retention policy (how many days/backups to keep)
- [ ] Backup filename format (date-based)
- [ ] Whether to compress the backup

### Script Configuration
- [ ] Cron schedule (e.g., `0 2 * * *` for 2 AM daily)
- [ ] Error handling/notification preferences
- [ ] Whether to verify backup integrity

## Implementation Steps
- [ ] Create backup script using `sqlite3 .backup` for consistent snapshots
- [ ] Use `aws s3 cp` for upload to S3
- [ ] Add cron job to run script nightly
- [ ] Test backup and restore process
- [ ] Set up monitoring/alerting for backup failures

## Database Location
- Database file: `/var/lib/link-shortener/database.db`
- Based on `DATABASE_PATH` environment variable in deployment