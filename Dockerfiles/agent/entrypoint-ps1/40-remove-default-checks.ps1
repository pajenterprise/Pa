if (-not (Test-Path env:DD_REMOVE_DEFAULT_CHECKS)) {
    exit 0
}

# TODO: remove all .yaml.default recursively, even though most are removed in 60-disable-infra.ps1
Remove-Item C:\ProgramData\Datadog\conf.d\cpu.d\conf.yaml.default
Remove-Item C:\ProgramData\Datadog\conf.d\memory.d\conf.yaml.default
Remove-Item C:\ProgramData\Datadog\conf.d\ntp.d\conf.yaml.default
