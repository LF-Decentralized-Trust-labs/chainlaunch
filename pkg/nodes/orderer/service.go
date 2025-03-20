package orderer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"
)

// startService starts the orderer as a system service
func (o *LocalOrderer) startService(cmd string, env map[string]string, dirPath string) (*StartServiceResponse, error) {
	platform := runtime.GOOS
	switch platform {
	case "linux":
		if err := o.createSystemdService(cmd, env, dirPath); err != nil {
			return nil, fmt.Errorf("failed to create systemd service: %w", err)
		}
		if err := o.startSystemdService(); err != nil {
			return nil, fmt.Errorf("failed to start systemd service: %w", err)
		}
		return &StartServiceResponse{
			Mode:        "service",
			Type:        "systemd",
			ServiceName: o.getServiceName(),
		}, nil

	case "darwin":
		if err := o.createLaunchdService(cmd, env, dirPath); err != nil {
			return nil, fmt.Errorf("failed to create launchd service: %w", err)
		}
		if err := o.startLaunchdService(); err != nil {
			return nil, fmt.Errorf("failed to start launchd service: %w", err)
		}
		return &StartServiceResponse{
			Mode:        "service",
			Type:        "launchd",
			ServiceName: o.getLaunchdServiceName(),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported platform for service mode: %s", platform)
	}
}

// createSystemdService creates a systemd service file
func (o *LocalOrderer) createSystemdService(cmd string, env map[string]string, dirPath string) error {
	var envStrings []string
	for k, v := range env {
		envStrings = append(envStrings, fmt.Sprintf("Environment=\"%s=%s\"", k, v))
	}

	tmpl := template.Must(template.New("systemd").Parse(`
[Unit]
Description=Hyperledger Fabric Orderer - {{.ID}}
After=network.target

[Service]
Type=simple
WorkingDirectory={{.DirPath}}
ExecStart={{.Cmd}}
Restart=on-failure
RestartSec=10
LimitNOFILE=65536
{{range .EnvVars}}{{.}}
{{end}}

[Install]
WantedBy=multi-user.target
`))

	data := struct {
		ID      string
		DirPath string
		Cmd     string
		EnvVars []string
	}{
		ID:      o.opts.ID,
		DirPath: dirPath,
		Cmd:     cmd,
		EnvVars: envStrings,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := os.WriteFile(o.getServiceFilePath(), buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write systemd service file: %w", err)
	}

	return nil
}

// createLaunchdService creates a launchd service file
func (o *LocalOrderer) createLaunchdService(cmd string, env map[string]string, dirPath string) error {
	var envStrings []string
	for k, v := range env {
		envStrings = append(envStrings, fmt.Sprintf("<key>%s</key>\n    <string>%s</string>", k, v))
	}

	tmpl := template.Must(template.New("launchd").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>{{.ServiceName}}</string>
  <key>ProgramArguments</key>
  <array>
      <string>/bin/bash</string>
      <string>-c</string>
      <string>{{.Cmd}}</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>StandardOutPath</key>
  <string>{{.LogPath}}</string>
  <key>StandardErrorPath</key>
  <string>{{.LogPath}}</string>
  <key>EnvironmentVariables</key>
  <dict>
    {{range .EnvVars}}{{.}}
    {{end}}
  </dict>
</dict>
</plist>`))

	data := struct {
		ServiceName string
		Cmd         string
		LogPath     string
		EnvVars     []string
	}{
		ServiceName: o.getLaunchdServiceName(),
		Cmd:         cmd,
		LogPath:     filepath.Join(dirPath, o.getServiceName()+".log"),
		EnvVars:     envStrings,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := os.WriteFile(o.getLaunchdPlistPath(), buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write launchd service file: %w", err)
	}

	return nil
}

// startSystemdService starts the systemd service
func (o *LocalOrderer) startSystemdService() error {
	if err := o.execSystemctl("daemon-reload"); err != nil {
		return err
	}
	if err := o.execSystemctl("enable", o.getServiceName()); err != nil {
		return err
	}
	if err := o.execSystemctl("start", o.getServiceName()); err != nil {
		return err
	}
	return o.execSystemctl("restart", o.getServiceName())
}

// startLaunchdService starts the launchd service
func (o *LocalOrderer) startLaunchdService() error {
	cmd := exec.Command("launchctl", "load", o.getLaunchdPlistPath())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to load launchd service: %w", err)
	}

	cmd = exec.Command("launchctl", "start", o.getLaunchdServiceName())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start launchd service: %w", err)
	}

	return nil
}

// stopSystemdService stops the systemd service
func (o *LocalOrderer) stopSystemdService() error {
	serviceName := o.getServiceName()

	// Stop the service
	if err := o.execSystemctl("stop", serviceName); err != nil {
		return fmt.Errorf("failed to stop systemd service: %w", err)
	}

	// Disable the service
	if err := o.execSystemctl("disable", serviceName); err != nil {
		o.logger.Warn("Failed to disable systemd service", "error", err)
	}

	// Remove the service file
	if err := os.Remove(o.getServiceFilePath()); err != nil {
		if !os.IsNotExist(err) {
			o.logger.Warn("Failed to remove service file", "error", err)
		}
	}

	// Reload systemd daemon
	if err := o.execSystemctl("daemon-reload"); err != nil {
		o.logger.Warn("Failed to reload systemd daemon", "error", err)
	}

	return nil
}

// stopLaunchdService stops the launchd service
func (o *LocalOrderer) stopLaunchdService() error {
	// Stop the service
	stopCmd := exec.Command("launchctl", "stop", o.getLaunchdServiceName())
	if err := stopCmd.Run(); err != nil {
		o.logger.Warn("Failed to stop launchd service", "error", err)
	}

	// Unload the service
	unloadCmd := exec.Command("launchctl", "unload", o.getLaunchdPlistPath())
	if err := unloadCmd.Run(); err != nil {
		return fmt.Errorf("failed to unload launchd service: %w", err)
	}

	return nil
}

// execSystemctl executes a systemctl command
func (o *LocalOrderer) execSystemctl(command string, args ...string) error {
	cmdArgs := append([]string{"systemctl", command}, args...)
	cmd := exec.Command("sudo", cmdArgs...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl %s failed: %w", command, err)
	}
	return nil
}

// startDocker starts the orderer in a docker container
func (o *LocalOrderer) startDocker(env map[string]string, mspConfigPath, dataConfigPath string) (*StartDockerResponse, error) {
	// TODO: Implement docker mode
	return nil, fmt.Errorf("docker mode not implemented")
}

// stopDocker stops the orderer docker container
func (o *LocalOrderer) stopDocker() error {
	// TODO: Implement docker mode
	return fmt.Errorf("docker mode not implemented")
}
