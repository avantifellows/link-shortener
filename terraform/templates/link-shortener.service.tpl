[Unit]
Description=Link Shortener Service
After=network.target
Wants=network.target

[Service]
Type=simple
User=${ssh_user}
Group=${ssh_user}
WorkingDirectory=/opt/link-shortener
ExecStart=/opt/link-shortener/link-shortener
EnvironmentFile=/opt/link-shortener/.env
Restart=always
RestartSec=5
TimeoutStopSec=20
KillMode=process

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/lib/link-shortener

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=link-shortener

[Install]
WantedBy=multi-user.target