# Troubleshooting Guide

This guide helps you resolve common issues with GitCells. If you can't find a solution here, please check our [GitHub Issues](https://github.com/Classic-Homes/gitcells/issues) page.

## Quick Fixes

Before diving into specific issues, try these quick fixes:

1. **Update GitCells**: `gitcells update`
2. **Check configuration**: `gitcells init --validate`
3. **Restart watchers**: Stop with Ctrl+C and start again
4. **Check permissions**: Ensure read/write access to directories
5. **View error logs**: `gitcells tui` → Error Logs

## Common Issues

### Installation Issues

#### "Command not found"

**Problem**: After installation, `gitcells` command is not recognized.

**Solutions**:
1. Check if GitCells is in your PATH:
   ```bash
   echo $PATH
   which gitcells
   ```

2. Add to PATH (Linux/macOS):
   ```bash
   echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
   source ~/.bashrc
   ```

3. Windows: Add GitCells directory to System PATH through Environment Variables

#### "Permission denied"

**Problem**: Can't execute GitCells binary.

**Solution**:
```bash
chmod +x /path/to/gitcells
```

### File Watching Issues

#### Watcher Not Detecting Changes

**Problem**: GitCells doesn't react when Excel files are saved.

**Solutions**:

1. **Check file extensions**:
   ```yaml
   # .gitcells.yaml
   watcher:
     file_extensions: [".xlsx", ".xls", ".xlsm"]
   ```

2. **Verify ignore patterns**:
   ```yaml
   watcher:
     ignore_patterns: ["~$*", "*.tmp"]
   ```

3. **Increase debounce delay**:
   ```yaml
   watcher:
     debounce_delay: 5s  # Increase from default 2s
   ```

4. **Check verbose output**:
   ```bash
   gitcells watch --verbose .
   ```

#### "Too many open files" Error

**Problem**: System limit reached when watching many files.

**Solutions**:

1. **Increase system limits** (macOS):
   ```bash
   ulimit -n 2048
   ```

2. **Linux**: Edit `/etc/security/limits.conf`:
   ```
   * soft nofile 4096
   * hard nofile 8192
   ```

3. **Watch fewer directories**:
   ```yaml
   watcher:
     directories: ["./active"]  # Not entire drive
   ```

### Conversion Issues

#### Excel File Won't Convert

**Problem**: Conversion fails with error message.

**Common Causes & Solutions**:

1. **File is open**:
   - Close Excel
   - Check Task Manager for Excel processes
   - Use `lsof | grep filename.xlsx` (Linux/macOS)

2. **Corrupted file**:
   - Try opening in Excel first
   - Save as new file
   - Use Excel's repair feature

3. **Unsupported features**:
   - Check error message for specific feature
   - Simplify the Excel file
   - Report issue on GitHub

4. **File too large**:
   ```yaml
   converter:
     chunking_strategy: "size-based"
     max_chunk_size: "5MB"
   ```

#### Formulas Not Preserved

**Problem**: Formulas missing or showing as values in JSON.

**Solutions**:

1. **Enable formula preservation**:
   ```bash
   gitcells convert file.xlsx --preserve-formulas
   ```

2. **Check configuration**:
   ```yaml
   converter:
     preserve_formulas: true
   ```

3. **Complex formulas**: Some formulas may need manual review

#### Styles/Formatting Lost

**Problem**: Cell formatting not preserved in conversion.

**Solutions**:

1. **Enable style preservation**:
   ```yaml
   converter:
     preserve_styles: true
   ```

2. **Check specific styles**: Some exotic styles may not be supported

### Git Integration Issues

#### "Not a git repository" Error

**Problem**: GitCells can't find Git repository.

**Solutions**:

1. **Initialize Git**:
   ```bash
   git init
   ```

2. **Check you're in the right directory**:
   ```bash
   pwd
   ls -la .git
   ```

#### Auto-commit Not Working

**Problem**: Changes aren't being committed automatically.

**Solutions**:

1. **Check configuration**:
   ```yaml
   git:
     auto_commit: true  # Not just in watcher section
   ```

2. **Verify Git setup**:
   ```bash
   git config user.name
   git config user.email
   ```

3. **Check Git status**:
   ```bash
   git status
   ```

#### Push Failed - Authentication

**Problem**: Can't push to remote repository.

**Solutions**:

1. **HTTPS authentication**:
   ```bash
   git config credential.helper store
   # Enter credentials once
   ```

2. **SSH authentication**:
   ```bash
   ssh-keygen -t rsa
   # Add public key to GitHub/GitLab
   ```

3. **Token authentication** (GitHub):
   ```bash
   git remote set-url origin https://TOKEN@github.com/user/repo.git
   ```

### Performance Issues

#### Slow Conversion

**Problem**: Large files take too long to convert.

**Solutions**:

1. **Enable chunking**:
   ```yaml
   converter:
     chunking_strategy: "size-based"
     max_chunk_size: "10MB"
   ```

2. **Reduce preserved features**:
   ```yaml
   converter:
     preserve_styles: false  # If not needed
     compact_json: true
   ```

3. **Increase worker threads**:
   ```yaml
   advanced:
     worker_threads: 8
   ```

#### High Memory Usage

**Problem**: GitCells uses too much RAM.

**Solutions**:

1. **Limit cell count**:
   ```yaml
   converter:
     max_cells_per_sheet: 500000
   ```

2. **Process files sequentially**:
   ```yaml
   advanced:
     worker_threads: 1
   ```

3. **Enable garbage collection**:
   ```bash
   export GOGC=50  # More aggressive GC
   ```

### Terminal UI Issues

#### Garbled Display

**Problem**: TUI shows incorrect characters or layout.

**Solutions**:

1. **Set UTF-8 locale**:
   ```bash
   export LANG=en_US.UTF-8
   export LC_ALL=en_US.UTF-8
   ```

2. **Check terminal**:
   ```bash
   echo $TERM
   # Should be xterm-256color or similar
   ```

3. **Try different terminal**: iTerm2, Windows Terminal, etc.

#### Can't Navigate

**Problem**: Arrow keys don't work in TUI.

**Solutions**:

1. **Use alternative keys**: j/k for up/down
2. **Check terminal settings**: Disable application keypad mode
3. **Try different terminal emulator**

### Error Messages Explained

#### "File locked by another process"

**Meaning**: Excel or another program has the file open.

**Fix**: Close all programs using the file.

#### "Checksum mismatch"

**Meaning**: File was modified during conversion.

**Fix**: Wait for file operations to complete.

#### "Invalid file format"

**Meaning**: File isn't a valid Excel file.

**Fix**: Verify file isn't corrupted, check extension matches content.

#### "Context deadline exceeded"

**Meaning**: Operation timed out.

**Fix**: Increase timeouts or reduce file size.

## Debug Mode

Enable debug logging for detailed information:

```bash
# Via command line
gitcells --verbose watch .

# Via configuration
advanced:
  log_level: "debug"

# Via environment
export GITCELLS_LOG_LEVEL=debug
```

## Getting More Help

### Collect Debug Information

When reporting issues, include:

1. **Version info**:
   ```bash
   gitcells version
   ```

2. **Configuration**:
   ```bash
   cat .gitcells.yaml
   ```

3. **Error logs**:
   ```bash
   gitcells tui  # → Error Logs → Export
   ```

4. **System info**:
   ```bash
   # OS version
   uname -a
   
   # Git version
   git --version
   
   # Disk space
   df -h
   ```

### Reporting Issues

1. Check [existing issues](https://github.com/Classic-Homes/gitcells/issues)
2. Create new issue with:
   - Clear problem description
   - Steps to reproduce
   - Error messages
   - Debug information
   - Expected vs actual behavior

### Community Support

- GitHub Discussions
- Stack Overflow tag: `gitcells`
- Email: support@gitcells.io

## Platform-Specific Issues

### Windows

- **Path separators**: Use forward slashes in config files
- **Long paths**: Enable long path support in Windows 10+
- **Antivirus**: Add GitCells to exclusions

### macOS

- **Gatekeeper**: `xattr -d com.apple.quarantine gitcells`
- **File watching limits**: Increase with `ulimit`
- **Permissions**: Grant Full Disk Access if needed

### Linux

- **SELinux**: May need policy adjustments
- **Inotify limits**: Increase system limits
- **Snap packages**: Use classic confinement

## Next Steps

- Review [Configuration](configuration.md) options
- Check [Command Reference](../reference/commands.md)
- Browse [GitHub Discussions](https://github.com/Classic-Homes/gitcells/discussions) for more help
- Join our [Community](https://github.com/Classic-Homes/gitcells/discussions)