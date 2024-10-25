# iShinobu
Ishinobu is a modular triage data collection tool for macOS intended to be used by incident responders, security analysts, and forensic investigators.
It is designed to collect system information and logs from various sources, such as system logs, network connections, running processes, and more.
Main features include:
- The collected data can be exported in JSON or CSV format.
- All events are logged to a file, and the output is compressed into a single file for easy sharing.
- Logs are timestamped under the same key, which is useful for correlating events across different sources, or just to have a chronological view of the collected data.
- The tool is modular, which means that new data collection modules can be easily added.
- The tool is designed to be run in parallel, which makes it faster to collect data from multiple sources.
- The tool is designed to be run in a macOS environment, but it can be easily adapted to other platforms.
- Developers have a way to create a template for new modules.

## Installation
```bash
go install github.com/gnzdotmx/ishinobu/ishinobu/cmd
```
## Usage
```bash
ishinobu -m all -e json -p 4 -v 1
```
### Verbosity Levels

The application supports two verbosity levels:

- `1`: Info and Error
- `2`: Debug, Info, and Error

## Modules
- **auditlogs**: Collects information from the macOS audit logs.
- **netstat**: Collects information about current network connections.
- **nettop**: Collects the amount of data transferred by processes and network interfaces.
- **ps**: Collects the list of running processes and their details.
- **unifiedlog**: Collects information from the macOS unified logs.
	- [Enabled] Command line activity - Run with elevated privileges.
	- [Enabled] SSH activity - Remmote connections.
	- [Enabled] Screen sharing activity - Remote desktop connections.
	- [Enabled] Session creation or deletion.
	- [Disabled] System logs - Kernel messages.
	- [Disabled] Security logs - Authentication attempts.
	- [Disabled] Network logs - Network activities.
	- [Disabled] User activity logs - Login sessions.
	- [Disabled] File system events - Disk mounts.
	- [Disabled] Configuration changes - Software installations.
	- [Disabled] Hardware events - Peripheral connections.
	- [Disabled] Time and date changes - System time adjustments.

# Guide for developers
- [DEV.md](./DEV.md)
