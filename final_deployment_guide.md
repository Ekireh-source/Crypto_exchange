# 🚀 Final Validated Deployment Guide
## Crypto Exchange — Next.js + Go + PostgreSQL + Redis + Nginx

This guide documents the exact, tested procedures we followed to successfully deploy the application on a **1GB DigitalOcean Droplet**, including all bug fixes for memory limitations, `.env` parsing, and systemd Node.js path issues.

---

## Step 1 — Overcoming the 1GB RAM Limit (Swap File)

Next.js `pnpm build` requires up to 2GB of memory and will instantly crash with a `SIGKILL` on a 1GB droplet. We solved this by adding a **2GB Swap File** (virtual memory on the hard drive).

```bash
# 1. Create a 2GB swap file
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# 2. Make it permanent across reboots
sudo cp /etc/fstab /etc/fstab.bak
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# 3. Optimize Linux to prioritize physical RAM over swap
sudo sysctl vm.swappiness=10
sudo sysctl vm.vfs_cache_pressure=50
echo 'vm.swappiness=10' | sudo tee -a /etc/sysctl.conf
echo 'vm.vfs_cache_pressure=50' | sudo tee -a /etc/sysctl.conf
```

---

## Step 2 — Backend Configuration & Build

The Go environment parser crashes if there are inline comments (`# ...`) next to numbers in the `.env` file. 

**1. Create the backend `.env` without inline comments:**
```bash
nano /var/www/Crypto_exchange/backend/.env
```
Ensure values look like this:
```env
EXCHANGE_FEE_RATE=0.001
REFERRAL_FEE_SHARE=0.30
```

**2. Fix Migration Order & Build:**
```bash
cd /var/www/Crypto_exchange/backend

# Fix the conflicting 003 migration files
mv internal/db/migrations/003_sweeper_tracking.sql internal/db/migrations/013_sweeper_tracking.sql

# Compile the Go binary
go build -o bin/exchange .

# Run DB Migrations
./bin/exchange migrate
```

---

## Step 3 — Frontend Configuration & Build

**1. Set the correct API URL:**
```bash
nano /var/www/Crypto_exchange/client/.env.production
```
```env
NEXT_PUBLIC_API_URL=http://busywt.site/v1
```

**2. Install and Build (safely using our new Swap file):**
```bash
cd /var/www/Crypto_exchange/client
pnpm install --frozen-lockfile
pnpm build
```

---

## Step 4 — Systemd Services (The "Bullet-proof" Method)

To keep the servers running forever in the background, we created two systemd services.

### Backend Service (Go)
```bash
sudo nano /etc/systemd/system/exchange-backend.service
```
```ini
[Unit]
Description=Crypto Exchange Backend (Go)
After=network.target postgresql.service redis-server.service

[Service]
Type=simple
User=oscar
Group=oscar
WorkingDirectory=/var/www/Crypto_exchange/backend
EnvironmentFile=/var/www/Crypto_exchange/backend/.env
ExecStart=/var/www/Crypto_exchange/backend/bin/exchange
Restart=always
RestartSec=5s
SyslogIdentifier=exchange-backend

[Install]
WantedBy=multi-user.target
```

### Frontend Service (Next.js via NVM)
*Note: Because `pnpm` and `.bin/next` bash shims fail inside systemd's restricted environment, we bypass them entirely by executing the Next.js javascript file directly using the absolute Node path.*

```bash
sudo nano /etc/systemd/system/exchange-frontend.service
```
```ini
[Unit]
Description=Crypto Exchange Frontend (Next.js)
After=network.target

[Service]
Type=simple
User=oscar
Group=oscar
WorkingDirectory=/var/www/Crypto_exchange/client
Environment=NODE_ENV=production
Environment=PORT=3000
EnvironmentFile=/var/www/Crypto_exchange/client/.env.production

# The exact absolute path to Node executing the exact path to Next.js
ExecStart=/home/oscar/.nvm/versions/node/v22.23.2/bin/node /var/www/Crypto_exchange/client/node_modules/next/dist/bin/next start -p 3000

Restart=always
RestartSec=5s
SyslogIdentifier=exchange-frontend

[Install]
WantedBy=multi-user.target
```

**Enable and Start Both:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable exchange-backend exchange-frontend
sudo systemctl start exchange-backend exchange-frontend
```

---

## Step 5 — Unified Nginx Reverse Proxy

Instead of using an `api.` subdomain, we route traffic perfectly using URL paths on a single domain.

```bash
sudo nano /etc/nginx/sites-available/exchange
```
```nginx
server {
    listen 80;
    server_name 157.230.98.105 busywt.site www.busywt.site;

    # 1. Route /v1/ API traffic to the Go backend (Port 8080)
    location /v1/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        
        proxy_read_timeout 120s;
        client_max_body_size 10M;
    }

    # 2. Route everything else to the Next.js frontend (Port 3000)
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 3. Cache Next.js static files aggressively
    location /_next/static/ {
        proxy_pass http://127.0.0.1:3000;
        add_header Cache-Control "public, max-age=31536000, immutable";
    }
}
```

**Enable Nginx Configuration:**
```bash
sudo ln -s /etc/nginx/sites-available/exchange /etc/nginx/sites-enabled/
sudo rm /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx
```

---

## Step 6 — DNS Configuration Checklist

Ensure your domain registrar has these exact `A` records pointing to your droplet IP. It may take up to 24 hours to propagate across all global wifi networks.

| Type | Host / Name | Value (IP) |
|---|---|---|
| `A` | `@` | `157.230.98.105` |
| `A` | `www` | `157.230.98.105` |

**To test if DNS is ready from your laptop:**
```bash
ping busywt.site
```
*(If it returns your IP address, it is live and ready!)*
