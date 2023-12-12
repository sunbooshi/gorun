# Simple Command Execution Service

This is a simple command execution service written in Go. It allows you to execute commands via HTTP requests and provides endpoints for checking executed commands and statistics.

## Functionality

- **Execute Command:** Send a POST request to `/run` with the command in the request body to execute the command.
- **Get Executed Commands Statistics:** Send a GET request to `/stats` to get statistics on executed commands.
- **Server Information:** Send a request to `/` to get information about available endpoints.

## Deployment

1. **Clone the Repository:**
   ```bash
   git clone https://github.com/sunbooshi/gorun.git
   cd gorun
   ```

2. **Configure Whitelist:**
   Edit the `/usr/local/etc/gorun.conf` file to specify the whitelist of allowed commands, with one command per line.

3. **Build and Run:**
   ```bash
   go build
   ./gorun -port=8080
   ```
   Replace `8080` with the desired port number.

4. **Access the Service:**
   The service will be accessible at `http://localhost:8080`. You can send HTTP requests to the specified endpoints.

## Permissions

### Log File Permissions

If you encounter permission issues with the log file (`/var/log/gorun.log`), ensure that the user running the service has the necessary permissions to write to this file. You may need to create the log file manually and adjust its permissions.

Example:
```bash
touch /var/log/gorun.log
chmod 666 /var/log/gorun.log
```

### Port Binding

If you choose a port number below 1024 (e.g., 80), you might need superuser privileges to bind to that port. Consider using a port number above 1024 or running the service with elevated privileges if required.

## Contributors

- ChatGPT
