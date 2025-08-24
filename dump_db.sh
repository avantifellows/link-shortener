#!/bin/bash
set -e

# Load deployment configuration (same as deploy.sh)
if [ -f ".env.deploy.local" ]; then
    echo "📋 Loading local deployment config..."
    source .env.deploy.local
elif [ -f ".env.deploy" ]; then
    echo "📋 Loading deployment config..."
    source .env.deploy
else
    echo "❌ Error: No deployment configuration found!"
    echo "   Please create .env.deploy.local with your server details"
    echo "   You can copy .env.deploy as a template"
    exit 1
fi

# Validate required variables
if [ -z "$SERVER" ] || [ -z "$KEY_PATH" ] || [ -z "$DEPLOY_DIR" ] || [ -z "$SERVICE_NAME" ]; then
    echo "❌ Error: Missing required deployment configuration!"
    echo "   Required: SERVER, KEY_PATH, DEPLOY_DIR, SERVICE_NAME"
    exit 1
fi
LOCAL_DB="./link_shortener.db"
REMOTE_DB_DIR="/var/lib/link-shortener"
REMOTE_DB="$REMOTE_DB_DIR/database.db"

echo "📦 Dumping local database to server..."

# Check if local database exists
if [ ! -f "$LOCAL_DB" ]; then
    echo "❌ Error: Local database not found at $LOCAL_DB"
    echo "   Make sure you have run the application locally and created some links."
    exit 1
fi

echo "📊 Local database info:"
echo "   Path: $LOCAL_DB"
echo "   Size: $(du -h $LOCAL_DB | cut -f1)"
echo "   Records: $(sqlite3 $LOCAL_DB 'SELECT COUNT(*) FROM link_mappings;') links"

echo ""
read -p "🤔 This will replace the server database. Continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Cancelled."
    exit 1
fi

# Stop the service to safely replace database
echo "⏸️  Stopping link shortener service..."
ssh -i $KEY_PATH $SERVER "sudo systemctl stop $SERVICE_NAME"

# Create backup of existing database (if it exists)
echo "💾 Creating backup of existing database..."
ssh -i $KEY_PATH $SERVER "
    if [ -f $REMOTE_DB ]; then 
        sudo cp $REMOTE_DB ${REMOTE_DB}.backup.$(date +%Y%m%d_%H%M%S)
        echo '   Backup created: ${REMOTE_DB}.backup.$(date +%Y%m%d_%H%M%S)'
    else
        echo '   No existing database found, skipping backup.'
    fi
"

# Ensure remote database directory exists with correct permissions
echo "📁 Ensuring database directory exists..."
ssh -i $KEY_PATH $SERVER "sudo mkdir -p $REMOTE_DB_DIR && sudo chown ubuntu:ubuntu $REMOTE_DB_DIR"

# Upload local database
echo "📤 Uploading local database..."
scp -i $KEY_PATH $LOCAL_DB $SERVER:/tmp/database.db
ssh -i $KEY_PATH $SERVER "sudo mv /tmp/database.db $REMOTE_DB && sudo chown ubuntu:ubuntu $REMOTE_DB"

# Start the service
echo "🏁 Starting link shortener service..."
ssh -i $KEY_PATH $SERVER "sudo systemctl start $SERVICE_NAME"

# Wait a moment for service to start
sleep 3

# Check service status
echo "✅ Checking service status..."
ssh -i $KEY_PATH $SERVER "sudo systemctl status $SERVICE_NAME --no-pager -l"

# Verify database was loaded correctly
echo "📊 Remote database info:"
ssh -i $KEY_PATH $SERVER "
    echo '   Path: $REMOTE_DB'
    echo '   Size: '\$(du -h $REMOTE_DB | cut -f1)
    echo '   Records: '\$(sqlite3 $REMOTE_DB 'SELECT COUNT(*) FROM link_mappings;')' links'
"

echo ""
echo "🎉 Database dump complete!"
echo "🌐 Your local data is now live at: https://temp-lnk.avantifellows.org"
echo "📊 Check the dashboard: https://temp-lnk.avantifellows.org/dashboard"