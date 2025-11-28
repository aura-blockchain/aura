# AURA Security Hardening Guide

**Version:** 1.0
**Last Updated:** 2025-11-25
**Target Audience:** Security Engineers, DevOps, System Administrators

---

## Table of Contents

1. [Overview](#overview)
2. [OS Hardening](#os-hardening)
3. [Network Security](#network-security)
4. [Key Management](#key-management)
5. [Access Control](#access-control)
6. [Audit Logging](#audit-logging)
7. [DDoS Protection](#ddos-protection)
8. [Incident Response](#incident-response)
9. [Security Monitoring](#security-monitoring)
10. [Compliance](#compliance)

---

## Overview

This guide provides enterprise-grade security hardening procedures for AURA blockchain nodes. Following these practices reduces attack surface and protects against common threats.

### Security Principles

1. **Defense in Depth**: Multiple layers of security
2. **Least Privilege**: Minimal permissions required
3. **Separation of Duties**: Role-based access control
4. **Regular Audits**: Continuous security assessment
5. **Incident Preparedness**: Documented response procedures

### Threat Model

**Threats:**
- **External Attackers**: DDoS, exploitation, unauthorized access
- **Insider Threats**: Malicious operators, credential theft
- **Supply Chain**: Compromised dependencies, malicious code
- **Physical**: Data center breach, hardware theft
- **Network**: Man-in-the-middle, eavesdropping

---

## OS Hardening

### Ubuntu 22.04 LTS Baseline

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Enable automatic security updates
sudo apt install -y unattended-upgrades
sudo dpkg-reconfigure -plow unattended-upgrades

# Configure automatic updates
sudo tee /etc/apt/apt.conf.d/50unattended-upgrades > /dev/null <<EOF
Unattended-Upgrade::Allowed-Origins {
    "\${distro_id}:\${distro_codename}-security";
};
Unattended-Upgrade::AutoFixInterruptedDpkg "true";
Unattended-Upgrade::MinimalSteps "true";
Unattended-Upgrade::Remove-Unused-Kernel-Packages "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
Unattended-Upgrade::Automatic-Reboot "false";
EOF
```

### Disable Unnecessary Services

```bash
# List running services
systemctl list-units --type=service --state=running

# Disable unnecessary services
sudo systemctl disable bluetooth.service
sudo systemctl disable cups.service
sudo systemctl disable avahi-daemon.service

# Remove unnecessary packages
sudo apt purge -y snapd
sudo apt autoremove -y
```

### Kernel Hardening

```bash
sudo tee /etc/sysctl.d/99-aura-security.conf > /dev/null <<EOF
# IP Forwarding (disable if not needed)
net.ipv4.ip_forward = 0
net.ipv6.conf.all.forwarding = 0

# Syncookies (SYN flood protection)
net.ipv4.tcp_syncookies = 1

# Ignore ICMP redirects
net.ipv4.conf.all.accept_redirects = 0
net.ipv6.conf.all.accept_redirects = 0
net.ipv4.conf.all.secure_redirects = 0

# Ignore source-routed packets
net.ipv4.conf.all.accept_source_route = 0
net.ipv6.conf.all.accept_source_route = 0

# Log martians (impossible addresses)
net.ipv4.conf.all.log_martians = 1

# Ignore ICMP ping requests (optional)
net.ipv4.icmp_echo_ignore_all = 0

# Protect against time-wait assassination
net.ipv4.tcp_rfc1337 = 1

# Enable random address generation
net.ipv6.conf.all.use_tempaddr = 2

# Disable IPv6 (if not used)
# net.ipv6.conf.all.disable_ipv6 = 1

# Kernel hardening
kernel.kptr_restrict = 2
kernel.dmesg_restrict = 1
kernel.yama.ptrace_scope = 2
kernel.unprivileged_bpf_disabled = 1
net.core.bpf_jit_harden = 2

# File system hardening
fs.protected_hardlinks = 1
fs.protected_symlinks = 1
fs.suid_dumpable = 0
EOF

sudo sysctl -p /etc/sysctl.d/99-aura-security.conf
```

### Filesystem Security

```bash
# Mount points with security options
# Edit /etc/fstab

# /tmp with noexec, nosuid, nodev
sudo tee -a /etc/fstab > /dev/null <<EOF
tmpfs /tmp tmpfs defaults,noexec,nosuid,nodev 0 0
tmpfs /var/tmp tmpfs defaults,noexec,nosuid,nodev 0 0
EOF

# Remount with new options
sudo mount -o remount /tmp
sudo mount -o remount /var/tmp

# Set proper permissions on critical files
sudo chmod 600 /boot/grub/grub.cfg
sudo chmod 644 /etc/passwd
sudo chmod 600 /etc/shadow
sudo chmod 644 /etc/group
sudo chmod 600 /etc/gshadow
```

---

## Network Security

### Firewall Configuration (UFW)

```bash
# Install UFW
sudo apt install -y ufw

# Default policies
sudo ufw default deny incoming
sudo ufw default allow outgoing

# SSH (limit rate)
sudo ufw limit 22/tcp comment 'SSH rate limited'

# For validator (private):
# Only allow from sentry nodes
sudo ufw allow from 10.0.2.1 to any port 26656 proto tcp comment 'Sentry 1'
sudo ufw allow from 10.0.3.1 to any port 26656 proto tcp comment 'Sentry 2'

# For sentry/full node (public):
sudo ufw allow 26656/tcp comment 'P2P'
sudo ufw allow 26657/tcp comment 'RPC'
sudo ufw allow 1317/tcp comment 'REST API'
sudo ufw allow 9090/tcp comment 'gRPC'

# Prometheus (only from monitoring server)
sudo ufw allow from MONITORING_IP to any port 26660 proto tcp comment 'Prometheus'

# Enable firewall
sudo ufw enable

# Verify rules
sudo ufw status numbered
```

### iptables Advanced Rules

```bash
#!/bin/bash
# advanced-firewall.sh - More granular control

# Flush existing rules
iptables -F
iptables -X

# Default policies
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Allow loopback
iptables -A INPUT -i lo -j ACCEPT

# Allow established connections
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# Rate limit SSH
iptables -A INPUT -p tcp --dport 22 -m state --state NEW -m recent --set
iptables -A INPUT -p tcp --dport 22 -m state --state NEW -m recent --update --seconds 60 --hitcount 4 -j DROP
iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# P2P with connection limit
iptables -A INPUT -p tcp --dport 26656 -m connlimit --connlimit-above 100 -j DROP
iptables -A INPUT -p tcp --dport 26656 -j ACCEPT

# RPC with rate limit
iptables -A INPUT -p tcp --dport 26657 -m limit --limit 10/sec --limit-burst 20 -j ACCEPT
iptables -A INPUT -p tcp --dport 26657 -j DROP

# Drop invalid packets
iptables -A INPUT -m state --state INVALID -j DROP

# Log dropped packets
iptables -A INPUT -m limit --limit 5/min -j LOG --log-prefix "iptables_INPUT_denied: " --log-level 7

# Save rules
iptables-save > /etc/iptables/rules.v4
```

### Fail2ban Configuration

```bash
# Install fail2ban
sudo apt install -y fail2ban

# SSH protection
sudo tee /etc/fail2ban/jail.local > /dev/null <<EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5
destemail = security@yourcompany.com
sendername = Fail2Ban
action = %(action_mwl)s

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 86400

[aurad-rpc]
enabled = true
port = 26657
filter = aurad-rpc
logpath = /var/log/nginx/access.log
maxretry = 100
findtime = 60
bantime = 600
EOF

# Create aurad-rpc filter
sudo tee /etc/fail2ban/filter.d/aurad-rpc.conf > /dev/null <<EOF
[Definition]
failregex = ^<HOST> .* "POST /broadcast_tx.*" 429
ignoreregex =
EOF

sudo systemctl restart fail2ban
sudo systemctl enable fail2ban
```

### VPN/Wireguard for Private Network

```bash
# Install Wireguard
sudo apt install -y wireguard

# Generate keys
wg genkey | tee privatekey | wg pubkey > publickey

# Configure server (validator)
sudo tee /etc/wireguard/wg0.conf > /dev/null <<EOF
[Interface]
Address = 10.0.1.1/24
ListenPort = 51820
PrivateKey = YOUR_PRIVATE_KEY

# Sentry 1
[Peer]
PublicKey = SENTRY1_PUBLIC_KEY
AllowedIPs = 10.0.2.1/32

# Sentry 2
[Peer]
PublicKey = SENTRY2_PUBLIC_KEY
AllowedIPs = 10.0.3.1/32
EOF

# Start Wireguard
sudo wg-quick up wg0
sudo systemctl enable wg-quick@wg0
```

---

## Key Management

### Validator Key Security

```bash
# Generate keys on air-gapped machine
# NEVER generate on internet-connected machine

# Encrypt keys with GPG
gpg --symmetric --cipher-algo AES256 priv_validator_key.json

# Create encrypted USB backup
sudo cryptsetup luksFormat /dev/sdX
sudo cryptsetup luksOpen /dev/sdX aura-backup
sudo mkfs.ext4 /dev/mapper/aura-backup
sudo mount /dev/mapper/aura-backup /mnt
sudo cp priv_validator_key.json.gpg /mnt/
sudo umount /mnt
sudo cryptsetup luksClose aura-backup
```

### Hardware Security Module (HSM)

**YubiHSM 2 Setup:**

```bash
# Install dependencies
sudo apt install -y libusb-1.0-0-dev

# Install Tendermint KMS
cargo install tmkms

# Initialize KMS
tmkms init /etc/tmkms

# Configure YubiHSM
sudo tee /etc/tmkms/tmkms.toml > /dev/null <<EOF
[[chain]]
id = "aura-mainnet-1"
key_format = { type = "bech32", account_key_prefix = "aurapub", consensus_key_prefix = "auravalconspub" }

[[validator]]
addr = "tcp://10.0.1.1:26658"
chain_id = "aura-mainnet-1"
reconnect = true
secret_key = "/etc/tmkms/secrets/kms-identity.key"

[[providers.yubihsm]]
adapter = { type = "usb" }
auth = { key = 1, password_file = "/etc/tmkms/secrets/password" }
keys = [{ chain_ids = ["aura-mainnet-1"], key = 1 }]
serial_number = "YOUR_SERIAL"
EOF

# Import validator key
tmkms yubihsm keys import -i 1 priv_validator_key.json

# Start KMS
tmkms start -c /etc/tmkms/tmkms.toml
```

### Key Rotation

```bash
# Generate new validator key
aurad init new-validator --chain-id aura-mainnet-1

# Submit key rotation transaction
aurad tx staking edit-validator \
  --new-pubkey=$(aurad tendermint show-validator) \
  --from=validator-operator \
  --chain-id=aura-mainnet-1

# Securely destroy old key
shred -vfz -n 10 old_priv_validator_key.json
```

### Keyring Backend

```bash
# Use OS keyring (most secure for operator keys)
aurad keys add operator --keyring-backend os

# Or file-based with encryption
aurad keys add operator --keyring-backend file

# Test keyring (Linux with libsecret)
sudo apt install -y gnome-keyring libsecret-1-0

# Export/Import with encryption
aurad keys export operator --keyring-backend os | gpg --symmetric --armor > operator.asc
gpg --decrypt operator.asc | aurad keys import operator --keyring-backend os
```

---

## Access Control

### User Management

```bash
# Create dedicated aura user
sudo useradd -m -s /bin/bash -G sudo aura
sudo passwd aura

# Disable root login
sudo passwd -l root

# Set proper home directory permissions
sudo chmod 750 /home/aura
sudo chown -R aura:aura /home/aura/.aura
sudo chmod 700 /home/aura/.aura
sudo chmod 600 /home/aura/.aura/config/*
sudo chmod 600 /home/aura/.aura/data/priv_validator_state.json
```

### SSH Key-Based Authentication

```bash
# Generate SSH key (on your machine, not server)
ssh-keygen -t ed25519 -C "aura-admin@yourcompany.com"

# Copy to server
ssh-copy-id -i ~/.ssh/id_ed25519.pub aura@server-ip

# Disable password authentication
sudo sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config
sudo systemctl restart sshd

# Test before logging out!
ssh aura@server-ip
```

### Sudo Configuration

```bash
# Configure sudo with logging
sudo visudo

# Add audit logging
Defaults logfile="/var/log/sudo.log"
Defaults log_year, log_host, loglinelen=0

# Require password for sudo (disable NOPASSWD)
aura ALL=(ALL:ALL) ALL

# Or specific commands only
aura ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart aurad
aura ALL=(ALL) NOPASSWD: /usr/bin/journalctl
```

### Multi-Factor Authentication (MFA)

```bash
# Install Google Authenticator
sudo apt install -y libpam-google-authenticator

# Configure for user
google-authenticator

# Enable in PAM
sudo tee -a /etc/pam.d/sshd > /dev/null <<EOF
auth required pam_google_authenticator.so
EOF

# Configure SSH
sudo sed -i 's/ChallengeResponseAuthentication no/ChallengeResponseAuthentication yes/' /etc/ssh/sshd_config
sudo systemctl restart sshd
```

---

## Audit Logging

### Enable auditd

```bash
# Install auditd
sudo apt install -y auditd audispd-plugins

# Configure audit rules
sudo tee /etc/audit/rules.d/aura.rules > /dev/null <<EOF
# Monitor configuration changes
-w /home/aura/.aura/config/ -p wa -k aura_config_change

# Monitor key files
-w /home/aura/.aura/config/priv_validator_key.json -p rwxa -k validator_key_access
-w /home/aura/.aura/data/priv_validator_state.json -p wa -k validator_state_change

# Monitor systemd service
-w /etc/systemd/system/aurad.service -p wa -k aurad_service_change

# Monitor sudo usage
-w /etc/sudoers -p wa -k sudoers_changes
-w /etc/sudoers.d/ -p wa -k sudoers_changes

# Monitor authentication
-w /etc/ssh/sshd_config -p wa -k sshd_config_change
-w /var/log/auth.log -p wa -k auth_log

# Monitor network configuration
-w /etc/ufw/ -p wa -k firewall_change
-a exit,always -F arch=b64 -S socket -S connect -k network_connections
EOF

# Reload rules
sudo augenrules --load
sudo systemctl restart auditd

# Query audit logs
sudo ausearch -k aura_config_change
sudo aureport -au
```

### Centralized Logging

**rsyslog to remote server:**

```bash
# Configure rsyslog
sudo tee /etc/rsyslog.d/50-aura.conf > /dev/null <<EOF
# Send aurad logs to remote syslog
:programname, isequal, "aurad" @@syslog-server.yourcompany.com:514

# Send auth logs
:programname, isequal, "sshd" @@syslog-server.yourcompany.com:514
EOF

sudo systemctl restart rsyslog
```

**ELK Stack Integration:**

```bash
# Install Filebeat
wget -qO - https://artifacts.elastic.co/GPG-KEY-elasticsearch | sudo apt-key add -
echo "deb https://artifacts.elastic.co/packages/8.x/apt stable main" | sudo tee /etc/apt/sources.list.d/elastic-8.x.list
sudo apt update && sudo apt install -y filebeat

# Configure Filebeat
sudo tee /etc/filebeat/filebeat.yml > /dev/null <<EOF
filebeat.inputs:
- type: journald
  id: aurad-logs
  include_matches:
    - _SYSTEMD_UNIT=aurad.service

output.elasticsearch:
  hosts: ["https://elasticsearch.yourcompany.com:9200"]
  username: "elastic"
  password: "your_password"
  index: "aura-logs-%{+yyyy.MM.dd}"
EOF

sudo systemctl enable filebeat
sudo systemctl start filebeat
```

---

## DDoS Protection

### Rate Limiting (nginx)

```bash
# Install nginx
sudo apt install -y nginx

# Configure rate limiting
sudo tee /etc/nginx/nginx.conf > /dev/null <<EOF
http {
    # Define rate limit zones
    limit_req_zone \$binary_remote_addr zone=rpc_limit:10m rate=10r/s;
    limit_req_zone \$binary_remote_addr zone=api_limit:10m rate=20r/s;
    limit_conn_zone \$binary_remote_addr zone=conn_limit:10m;

    # Connection limits
    limit_conn conn_limit 10;

    # Include site configs
    include /etc/nginx/sites-enabled/*;
}
EOF

# Site configuration
sudo tee /etc/nginx/sites-available/aura-rpc > /dev/null <<EOF
upstream aura_rpc {
    server 127.0.0.1:26657;
}

server {
    listen 80;
    server_name rpc.yournode.com;

    location / {
        limit_req zone=rpc_limit burst=20 nodelay;
        limit_req_status 429;

        proxy_pass http://aura_rpc;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    }
}
EOF

sudo ln -s /etc/nginx/sites-available/aura-rpc /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl restart nginx
```

### CloudFlare Protection

```yaml
# CloudFlare configuration for public RPC
Security Level: High
Challenge Passage: 30 minutes

Rate Limiting Rules:
  - Path: /broadcast_tx*
    Rate: 10 requests per minute per IP
    Action: Block

  - Path: /*
    Rate: 100 requests per minute per IP
    Action: Challenge

Firewall Rules:
  - Block known bad bots
  - Challenge anonymous proxies/VPNs
  - Whitelist known good IPs
```

### Connection Limits

```toml
# config.toml
[p2p]
max_num_inbound_peers = 100
max_num_outbound_peers = 50

[rpc]
max_open_connections = 900
max_subscription_clients = 100
max_subscriptions_per_client = 5

[mempool]
size = 5000
max_txs_bytes = 1073741824
```

---

## Incident Response

### Incident Response Plan

```markdown
# AURA Incident Response Plan

## Phase 1: Detection
- Automated monitoring alerts
- Manual security reviews
- User reports
- Audit log analysis

## Phase 2: Containment
- Isolate affected systems
- Block malicious IPs
- Disable compromised accounts
- Stop services if necessary

## Phase 3: Investigation
- Collect evidence
- Review logs
- Identify attack vector
- Assess damage

## Phase 4: Remediation
- Patch vulnerabilities
- Rotate compromised credentials
- Restore from clean backups
- Update security controls

## Phase 5: Recovery
- Restore services
- Verify integrity
- Monitor for re-infection
- Communication to stakeholders

## Phase 6: Post-Incident
- Document lessons learned
- Update procedures
- Improve detection
- Training and awareness
```

### Emergency Contacts

```yaml
Incident Response Team:
  Primary: security@yourcompany.com, +1-xxx-xxx-xxxx
  Secondary: ops@yourcompany.com, +1-xxx-xxx-xxxx

External Contacts:
  AURA Core Team: security@aura.network
  Data Center: support@datacenter.com, +1-xxx-xxx-xxxx
  Law Enforcement: local-cybercrime-unit@police.gov

Escalation:
  Level 1 (Low): Email to security team
  Level 2 (Medium): Email + SMS to on-call
  Level 3 (High): Phone call + PagerDuty
  Level 4 (Critical): All channels + management
```

### Forensics Preservation

```bash
#!/bin/bash
# forensics-collect.sh

INCIDENT_DIR="/forensics/incident-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$INCIDENT_DIR"

# Capture system state
ps auxww > "$INCIDENT_DIR/processes.txt"
netstat -antup > "$INCIDENT_DIR/network.txt"
ss -antup > "$INCIDENT_DIR/sockets.txt"
iptables -L -n -v > "$INCIDENT_DIR/iptables.txt"

# Copy logs
cp -r /var/log "$INCIDENT_DIR/logs"
journalctl -u aurad > "$INCIDENT_DIR/aurad.log"
cp /var/log/auth.log "$INCIDENT_DIR/"
cp /var/log/syslog "$INCIDENT_DIR/"

# Audit logs
sudo ausearch -ts recent > "$INCIDENT_DIR/audit.log"

# Configuration
cp -r ~/.aura/config "$INCIDENT_DIR/"

# Memory dump (optional, if rootkit suspected)
# sudo dd if=/dev/mem of="$INCIDENT_DIR/memory.dump" bs=1M

# Create archive
tar -czf "$INCIDENT_DIR.tar.gz" "$INCIDENT_DIR"
sha256sum "$INCIDENT_DIR.tar.gz" > "$INCIDENT_DIR.tar.gz.sha256"

echo "Forensics collected: $INCIDENT_DIR.tar.gz"
```

---

## Security Monitoring

### OSSEC/Wazuh Integration

```bash
# Install Wazuh agent
wget https://packages.wazuh.com/4.x/apt/pool/main/w/wazuh-agent/wazuh-agent_4.7.0-1_amd64.deb
sudo dpkg -i wazuh-agent_4.7.0-1_amd64.deb

# Configure agent
sudo tee /var/ossec/etc/ossec.conf > /dev/null <<EOF
<ossec_config>
  <client>
    <server>
      <address>wazuh-manager.yourcompany.com</address>
    </server>
  </client>

  <localfile>
    <log_format>syslog</log_format>
    <location>/var/log/auth.log</location>
  </localfile>

  <localfile>
    <log_format>command</log_format>
    <command>journalctl -u aurad -n 100</command>
    <frequency>120</frequency>
  </localfile>
</ossec_config>
EOF

sudo systemctl start wazuh-agent
sudo systemctl enable wazuh-agent
```

### Security Metrics

```yaml
# Prometheus security metrics
metrics:
  - failed_ssh_attempts_total
  - firewall_blocked_packets_total
  - sudo_commands_total
  - config_changes_total
  - failed_aurad_transactions_total
  - anomaly_detections_total
```

---

## Compliance

### CIS Benchmarks

Follow CIS Ubuntu 22.04 LTS Benchmark:

```bash
# Download CIS-CAT Lite
wget https://workbench.cisecurity.org/files/3884

# Run assessment
sudo ./CIS-CAT.sh -b benchmarks/CIS_Ubuntu_Linux_22.04_LTS_Benchmark_v1.0.0-xccdf.xml

# Review report and remediate findings
```

### SOC 2 Controls

**For SOC 2 compliance:**

1. **Access Control (CC6.1-CC6.3)**
   - Multi-factor authentication
   - Least privilege access
   - Regular access reviews

2. **Logical and Physical Access (CC6.6-CC6.8)**
   - Strong passwords
   - Session timeouts
   - Physical security at datacenter

3. **System Operations (CC7.1-CC7.5)**
   - Change management
   - Monitoring and logging
   - Incident response

4. **Change Management (CC8.1)**
   - Documented procedures
   - Testing before deployment
   - Rollback capabilities

### GDPR Considerations

For AURA's compliance module:

```bash
# Data subject access request
aurad query compliance data-subject-info <address>

# Right to erasure
aurad tx compliance erase-personal-data <address> --from operator

# Audit trail
aurad query compliance audit-log <address>
```

---

**Document Status**: Production Ready
**Review Cycle**: Quarterly
**Next Review**: 2026-02-25
