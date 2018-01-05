#!/bin/bash
set -ex

if [[ $TRAVIS_OS_NAME == "linux" ]]; then
	sysctl -w net.ipv6.conf.lo.disable_ipv6=0
	sysctl -w net.ipv6.conf.default.disable_ipv6=0
	sysctl -w net.ipv6.conf.all.disable_ipv6=0;
else
	cat > /Library/LaunchDaemons/limit.maxfiles.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>limit.maxfiles</string>
    <key>ProgramArguments</key>
    <array>
      <string>launchctl</string>
      <string>limit</string>
      <string>maxfiles</string>
      <string>65535</string>
      <string>65535</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>ServiceIPC</key>
    <false/>
  </dict>
</plist>
EOF
	chown root:wheel /Library/LaunchDaemons/limit.maxfiles.plist
	launchctl load -w /Library/LaunchDaemons/limit.maxfiles.plist
	launchctl limit maxfiles
fi
