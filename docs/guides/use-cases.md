# Common Use Cases

Real-world examples of using GitCells in different scenarios.

## Financial Reporting

### Monthly Financial Close

**Scenario**: Finance team needs to track changes during monthly close process.

**Setup**:
```yaml
# .gitcells.yml
project:
  name: "Monthly Financial Close"

files:
  patterns:
    - "close/*.xlsx"
    - "reconciliation/*.xlsx"
  
watch:
  enabled: true
  auto_commit: true
  
hooks:
  post_convert:
    - name: "Validate balances"
      script: "./scripts/validate-balances.py"
```

**Workflow**:
```bash
# Start of month close
git checkout -b close-2024-01
gitcells watch

# Team works on files
# Changes auto-tracked

# Review all changes
gitcells report --since "month-start"

# Finalize and merge
git commit -m "January 2024 close completed"
git checkout main
git merge close-2024-01
```

### Budget Planning

**Scenario**: Annual budget planning with multiple departments.

**Structure**:
```
/budgets
  /2024
    /departments
      - sales-budget.xlsx
      - marketing-budget.xlsx
      - operations-budget.xlsx
    /consolidated
      - master-budget.xlsx
```

**Configuration**:
```yaml
team:
  permissions:
    sales_team:
      edit: ["budgets/*/departments/sales-*.xlsx"]
    marketing_team:
      edit: ["budgets/*/departments/marketing-*.xlsx"]
    
  locking:
    enabled: true
    auto_unlock_after: "4h"
```

**Process**:
```bash
# Each department works independently
gitcells lock budgets/2024/departments/sales-budget.xlsx
# Make changes
gitcells unlock budgets/2024/departments/sales-budget.xlsx

# Consolidate budgets
gitcells validate budgets/2024/departments/*.xlsx
python consolidate-budgets.py
git commit -m "Consolidated 2024 department budgets"
```

## Data Analysis

### Research Data Collection

**Scenario**: Research team collecting and analyzing experimental data.

**Setup**:
```yaml
# .gitcells.yml
conversion:
  json:
    include_metadata: true
    include_empty_cells: false
    
validation:
  rules:
    - name: "Data integrity"
      script: "./validate-research-data.py"
      required: true
    
audit:
  enabled: true
  include_cell_values: true
  retention: "5 years"
```

**Workflow**:
```bash
# Import new data
gitcells import raw-data.csv experiment-001.xlsx
gitcells validate experiment-001.xlsx

# Track analysis changes
gitcells watch --filter "formulas,values"

# Generate audit trail
gitcells audit --from 2024-01-01 --export audit-trail.pdf
```

### Sales Analytics

**Scenario**: Sales team tracking performance metrics.

**Dashboard Setup**:
```yaml
watch:
  patterns: ["dashboards/*.xlsx", "data/*.xlsx"]
  
  triggers:
    - name: "Update dashboard"
      condition: "file_match('data/*.xlsx')"
      action:
        script: "./update-dashboard.py"
        
    - name: "Alert on target miss"
      condition: "cell_value('Dashboard!B5') < cell_value('Dashboard!C5')"
      action:
        email: "sales-manager@company.com"
```

## Project Management

### Resource Planning

**Scenario**: IT team managing resource allocation across projects.

**Excel Structure**:
```
- resource-master.xlsx
  - Sheet: "Resources" (team members)
  - Sheet: "Projects" (active projects)
  - Sheet: "Allocation" (who works on what)
  - Sheet: "Timeline" (Gantt chart)
```

**Tracking Changes**:
```bash
# Daily standup - check changes
gitcells diff resource-master.xlsx @{yesterday}

# Weekly planning
git checkout -b week-45-planning
# Update allocations
gitcells watch
git commit -m "Week 45 resource allocation"

# Conflict resolution
gitcells resolve resource-master.xlsx --strategy "max" # Take higher allocation
```

### Status Reporting

**Scenario**: Weekly project status reports.

**Automation**:
```yaml
# .gitcells.yml
schedules:
  - name: "Weekly status snapshot"
    cron: "0 17 * * 5"  # Friday 5 PM
    actions:
      - convert: "status-reports/*.xlsx"
      - commit: "Weekly status snapshot"
      - tag: "status-{date}"
      - email:
          to: "stakeholders@company.com"
          subject: "Weekly Status Update"
          attach: ["status-summary.pdf"]
```

## Compliance and Audit

### SOX Compliance

**Scenario**: Maintaining SOX-compliant financial records.

**Configuration**:
```yaml
# .gitcells.yml
security:
  audit:
    enabled: true
    log_all_changes: true
    
  protection:
    remove_personal_info: true
    require_approval: true
    approvers: ["cfo@company.com", "controller@company.com"]
    
  retention:
    period: "7 years"
    archive_location: "s3://sox-archives/"
```

**Workflow**:
```bash
# Quarterly attestation
gitcells audit --quarter Q4-2023 --export sox-report.pdf

# Lock down period-end files
gitcells lock period-end/*.xlsx --message "Period closed"
gitcells protect period-end/*.xlsx --permanent
```

### Change Control

**Scenario**: Tracking all changes for regulatory review.

**Implementation**:
```yaml
hooks:
  pre_commit:
    - name: "Require change ticket"
      script: |
        if ! git diff --cached --name-only | grep -q "TICKET-[0-9]+"
        then
          echo "Commit message must include ticket number"
          exit 1
        fi
        
  post_commit:
    - name: "Log to audit system"
      script: "./log-to-audit.py"
      args: ["{commit_hash}", "{files_changed}"]
```

## Educational Settings

### Grade Tracking

**Scenario**: Teachers managing student grades.

**Setup**:
```
/grades
  /2024-spring
    - CS101-grades.xlsx
    - CS102-grades.xlsx
  /2024-fall
    - CS101-grades.xlsx
```

**Privacy Configuration**:
```yaml
security:
  protection:
    remove_personal_info: true
    anonymize_on_export: true
    
  encryption:
    enabled: true
    key_file: "~/.gitcells/grades.key"
```

**Workflow**:
```bash
# Import from LMS
gitcells import canvas-export.csv CS101-grades.xlsx

# Track changes during grading
gitcells watch --auto-commit

# Generate reports
gitcells export CS101-grades.xlsx --format pdf --anonymize

# Archive at semester end
gitcells archive 2024-spring/*.xlsx --compress
```

## Manufacturing

### Inventory Management

**Scenario**: Tracking inventory levels and orders.

**Real-time Sync**:
```yaml
integrations:
  database:
    type: "postgresql"
    connection: "${DB_CONNECTION}"
    sync_interval: "15m"
    
    mappings:
      - excel: "Inventory!A:F"
        table: "inventory_levels"
      - excel: "Orders!*"
        table: "purchase_orders"
```

**Alerts**:
```yaml
triggers:
  - name: "Low inventory alert"
    condition: |
      any_cell_in_range("Inventory!E:E") < 
      corresponding_cell_in_range("Inventory!F:F")
    action:
      email: "purchasing@company.com"
      sms: "+1234567890"
```

## Healthcare

### Patient Data Tracking

**Scenario**: Anonymized patient data analysis.

**HIPAA Compliance**:
```yaml
security:
  compliance: "HIPAA"
  
  anonymization:
    enabled: true
    fields: ["patient_name", "ssn", "dob"]
    method: "hash"
    
  audit:
    enabled: true
    log_access: true
    log_modifications: true
    
  encryption:
    at_rest: true
    in_transit: true
```

**Workflow**:
```bash
# Import and anonymize
gitcells import patient-data.csv --anonymize

# Analyze trends
gitcells watch analysis/*.xlsx

# Export for research
gitcells export analysis/*.xlsx --format csv --verify-anonymized
```

## Best Practices by Industry

### Financial Services
- Enable audit logging
- Use approval workflows
- Implement change control
- Regular backups
- Encrypt sensitive data

### Healthcare
- HIPAA compliance mode
- Anonymize patient data
- Audit all access
- Encrypted storage
- Access controls

### Education
- Student privacy protection
- Semester-based archiving
- Grade change tracking
- Export controls

### Manufacturing
- Real-time database sync
- Inventory alerts
- Supply chain tracking
- Quality control logs

### Research
- Data integrity validation
- Reproducibility tracking
- Collaboration controls
- Long-term archiving

## Integration Examples

### With CI/CD

```yaml
# .github/workflows/excel-validation.yml
name: Excel Validation
on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Validate Excel files
        run: |
          gitcells validate *.xlsx
          gitcells test-roundtrip *.xlsx
```

### With Slack

```yaml
notifications:
  slack:
    webhook: "${SLACK_WEBHOOK}"
    
    templates:
      change: "📊 {user} updated {filename}"
      conflict: "⚠️ Conflict in {filename}"
      error: "❌ Error processing {filename}: {error}"
```

### With Jira

```yaml
integrations:
  jira:
    url: "https://company.atlassian.net"
    project: "EXCEL"
    
    create_issue_on:
      - validation_error
      - merge_conflict
      - formula_error
```

## Next Steps

- Set up your specific use case
- Configure [team collaboration](collaboration.md)
- Implement [auto-sync](auto-sync.md)
- Review [best practices](../reference/configuration.md)