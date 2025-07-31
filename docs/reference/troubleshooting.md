# Troubleshooting Guide

Solutions for common GitCells issues.

## Installation Issues

### "Command not found"

**Problem**: `gitcells: command not found` after installation

**Solutions**:

1. Check if GitCells is in your PATH:
```bash
echo $PATH
which gitcells
```

2. Add to PATH manually:
```bash
# macOS/Linux
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Windows
# Add C:\Program Files\GitCells to System PATH
```

3. Verify installation:
```bash
ls -la /usr/local/bin/gitcells
gitcells --version
```

### Permission Denied

**Problem**: Permission denied when installing or running

**Solutions**:

```bash
# macOS/Linux
sudo chmod +x /usr/local/bin/gitcells
sudo chown $USER:$USER /usr/local/bin/gitcells

# If still issues
sudo gitcells init
sudo chown -R $USER:$USER .gitcells/
```

## Conversion Issues

### "File is corrupted"

**Problem**: Cannot convert Excel file, shows as corrupted

**Solutions**:

1. Try recovery mode:
```bash
gitcells convert file.xlsx --recover
```

2. Check file in Excel:
- Open in Excel
- Save As new file
- Try converting the new file

3. Use compatibility mode:
```bash
gitcells convert file.xlsx --compatibility
```

### "Out of memory"

**Problem**: Large files cause memory errors

**Solutions**:

1. Use streaming mode:
```bash
gitcells convert large.xlsx --stream
```

2. Increase memory limit:
```yaml
# .gitcells.yml
performance:
  memory:
    max_heap: "4GB"
    low_memory_mode: true
```

3. Convert specific sheets:
```bash
gitcells convert large.xlsx --sheets "Summary"
```

### "Formula errors"

**Problem**: Formulas not converting correctly

**Solutions**:

1. Validate formulas:
```bash
gitcells validate file.xlsx --check-formulas
```

2. Check external references:
```bash
gitcells validate file.xlsx --check-references
```

3. Force formula evaluation:
```yaml
# .gitcells.yml
conversion:
  cells:
    evaluate_formulas: true
```

## Sync Issues

### "Files out of sync"

**Problem**: Excel and JSON files don't match

**Solutions**:

1. Force sync:
```bash
gitcells sync --force --all
```

2. Check sync status:
```bash
gitcells status --detailed
```

3. Reset sync state:
```bash
gitcells cache clear
gitcells sync --reset
```

### "Sync loop detected"

**Problem**: Files keep syncing back and forth

**Solutions**:

1. Increase debounce time:
```yaml
# .gitcells.yml
watch:
  debounce: "10s"
```

2. Check for circular dependencies:
```bash
gitcells validate *.xlsx --check-circular
```

3. Disable auto-sync temporarily:
```bash
gitcells watch --no-auto-sync
```

## Watch Issues

### "Changes not detected"

**Problem**: File changes aren't being picked up

**Solutions**:

1. Check ignore patterns:
```bash
gitcells watch --debug
```

2. Use polling mode:
```yaml
# .gitcells.yml
watch:
  polling: true
  interval: "5s"
```

3. Check file permissions:
```bash
ls -la *.xlsx
```

4. Clear watch cache:
```bash
gitcells cache clear --watch
gitcells watch --reset
```

### "Too many files open"

**Problem**: Error about file handle limits

**Solutions**:

1. Increase system limits:
```bash
# macOS
ulimit -n 10000

# Linux
echo "fs.inotify.max_user_watches=524288" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

2. Reduce watched files:
```yaml
# .gitcells.yml
watch:
  patterns: ["important/*.xlsx"]
  ignore: ["temp/*", "backup/*"]
```

## Git Integration Issues

### "Merge conflicts"

**Problem**: Git merge conflicts in JSON files

**Solutions**:

1. Use GitCells conflict resolution:
```bash
gitcells conflict file.xlsx --resolve
```

2. Configure auto-resolution:
```yaml
# .gitcells.yml
sync:
  conflict_resolution:
    strategy: "newer"
```

3. Prevent conflicts:
```bash
gitcells lock file.xlsx
```

### "Large diffs"

**Problem**: Git diffs are too large to review

**Solutions**:

1. Use GitCells diff:
```bash
gitcells diff file.xlsx --summary
```

2. Configure diff options:
```yaml
# .gitcells.yml
git:
  diff:
    algorithm: "patience"
    context_lines: 3
```

3. Use diff filters:
```bash
gitcells diff file.xlsx --filter values --threshold 100
```

## Performance Issues

### "Slow conversion"

**Problem**: Conversion takes too long

**Solutions**:

1. Enable parallel processing:
```yaml
# .gitcells.yml
performance:
  parallel:
    enabled: true
    workers: 8
```

2. Use caching:
```yaml
cache:
  enabled: true
  ttl: "24h"
```

3. Profile performance:
```bash
gitcells convert file.xlsx --profile
```

### "High CPU usage"

**Problem**: GitCells uses too much CPU

**Solutions**:

1. Limit workers:
```yaml
performance:
  parallel:
    workers: 2
  cpu_limit: "50%"
```

2. Increase debounce:
```yaml
watch:
  debounce: "10s"
  batch_changes: true
```

3. Use nice priority:
```bash
nice -n 10 gitcells watch
```

## Network Issues

### "Cannot access network drive"

**Problem**: Network files not accessible

**Solutions**:

1. Mount drive properly:
```bash
# Windows
net use Z: \\server\share

# macOS
mount_smbfs //server/share /mnt/share

# Linux
mount -t cifs //server/share /mnt/share
```

2. Configure retry:
```yaml
# .gitcells.yml
watch:
  network:
    retry_attempts: 5
    retry_delay: "10s"
    offline_mode: true
```

### "Webhook failures"

**Problem**: Notifications not being sent

**Solutions**:

1. Test webhook:
```bash
gitcells test-webhook --url https://hooks.slack.com/...
```

2. Check logs:
```bash
gitcells logs --filter webhook
```

3. Configure timeout:
```yaml
notifications:
  timeout: "30s"
  retry: 3
```

## Common Error Messages

### "Excel file is locked"

**Cause**: File is open in Excel
**Solution**: Close Excel or use read-only mode:
```bash
gitcells convert file.xlsx --readonly
```

### "Invalid configuration"

**Cause**: Syntax error in .gitcells.yml
**Solution**: Validate configuration:
```bash
gitcells config --validate
```

### "Git repository not found"

**Cause**: Not in a Git repository
**Solution**: Initialize Git:
```bash
git init
gitcells init
```

### "Circular reference detected"

**Cause**: Circular formulas in Excel
**Solution**: Fix in Excel or ignore:
```bash
gitcells convert file.xlsx --ignore-circular
```

## Debugging

### Enable Debug Mode

```bash
# Verbose output
gitcells -v convert file.xlsx

# Debug logging
export GITCELLS_LOG_LEVEL=debug
gitcells watch

# Trace mode
gitcells --trace convert file.xlsx
```

### Check Logs

```bash
# View logs
tail -f .gitcells/logs/gitcells.log

# Filter errors
grep ERROR .gitcells/logs/gitcells.log

# Export logs
gitcells logs --export debug-logs.zip
```

### System Information

```bash
# Run diagnostics
gitcells doctor

# Show environment
gitcells debug --env

# Test installation
gitcells test --all
```

## Getting Help

### Resources

1. **Documentation**: Check other guides in `/docs`
2. **GitHub Issues**: https://github.com/Classic-Homes/gitcells/issues
3. **Community Forum**: https://forum.gitcells.com
4. **Email Support**: support@gitcells.com

### Reporting Issues

When reporting issues, include:

1. GitCells version: `gitcells --version`
2. Operating system: `uname -a` or Windows version
3. Git version: `git --version`
4. Excel version (if relevant)
5. Configuration file: `.gitcells.yml`
6. Error messages and logs
7. Steps to reproduce

### Debug Bundle

Create a debug bundle:
```bash
gitcells debug --bundle
```

This creates `gitcells-debug.zip` with:
- Configuration
- Recent logs
- System information
- Sample files (sanitized)

## Recovery

### Reset GitCells

```bash
# Full reset
gitcells reset --all

# Reset configuration
gitcells config --reset

# Reset cache
gitcells cache clear --all

# Reset watch state
gitcells watch --reset
```

### Backup and Restore

```bash
# Backup
gitcells backup --all

# Restore
gitcells restore --from backup-2024-01-15.tar.gz
```

## Next Steps

- Review [command reference](commands.md)
- Check [configuration options](configuration.md)
- Read [best practices](../guides/collaboration.md)