# simple-job-tracker-backend

Go Fiber backend for the simple job tracker app.

---

## Deploy to Ubuntu VPS

### 1. Set up the server

```bash
# Create directory for the app
sudo mkdir -p /opt/simple-job-tracker-backend

# Create a systemd service
sudo tee /etc/systemd/system/simple-job-tracker.service << 'EOF'
[Unit]
Description=Simple Job Tracker Backend
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/simple-job-tracker-backend
ExecStart=/opt/simple-job-tracker-backend/simple-job-tracker-backend
Restart=always
RestartSec=5
EnvironmentFile=/opt/simple-job-tracker-backend/.env

[Install]
WantedBy=multi-user.target
EOF

# Create .env file (replace with your values)
sudo tee /opt/simple-job-tracker-backend/.env << 'EOF'
DATABASE_URL=postgres://postgres:your_password@localhost:5432/job_tracker?sslmode=disable
JWT_SECRET=generate-a-random-secret-here
PORT=3000
EOF

# Secure the .env file
sudo chmod 600 /opt/simple-job-tracker-backend/.env

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable simple-job-tracker
sudo systemctl start simple-job-tracker

# Check status
sudo systemctl status simple-job-tracker
```

### 2. Set up GitHub secrets

Go to your repo → Settings → Secrets and variables → Actions → Add these secrets:

| Secret         | Value                                          |
|----------------|------------------------------------------------|
| `VPS_HOST`     | Your VPS IP address or hostname                |
| `VPS_USER`     | SSH username (e.g. `root`)                     |
| `VPS_SSH_KEY`  | Private SSH key (e.g. contents of `~/.ssh/id_rsa`) |

### 3. Deploy

Push to `main` — the workflow in `.github/workflows/deploy.yml` will:
1. Build the Go binary for linux/amd64
2. Copy it to `/opt/simple-job-tracker-backend/` on your VPS
3. Restart the `simple-job-tracker` systemd service
