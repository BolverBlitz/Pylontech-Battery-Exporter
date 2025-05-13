# Install
Create a .env file in the root directory. This file should contain the following variables:  
```env
LOG_VERBOSE=false
DEVICE_IP=DeviceIP
REFRESH_SECONDS=30
PORT=9092
PROM_NAMESPACE=devicemon
```

Download the latest release of the exporter and mark it as executable:  
```bash
# Download the latest release of the exporter
ARCH=$(uname -m); OS=$(uname -s | tr '[:upper:]' '[:lower:]'); \
case $ARCH in x86_64) ARCH=amd64;; i386|i686) ARCH=386;; aarch64) ARCH=arm64;; esac; \
FILENAME="pylontech-prom-export-${OS}-${ARCH}"; \
wget -q "https://github.com/BolverBlitz/Pylontech-Battery-Exporter/releases/latest/download/${FILENAME}" -O ${FILENAME}

# Mark the file as executable
chmod +x pylontech-prom-export-*
```