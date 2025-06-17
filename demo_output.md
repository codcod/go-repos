# Enhanced Health Report Demo

## Key Improvements Made

### 1. Repository-Separated Reports
- Each repository now gets its own dedicated section with clear visual borders
- Eliminates confusion when analyzing multiple repositories
- Professional boxed layout for better readability

### 2. Enhanced Cyclomatic Complexity Analysis

#### Overall Metrics:
- **Total Complexity**: Sum of all function complexities
- **Maximum Complexity**: Highest complexity function found
- **Average Complexity**: Mean complexity across all functions
- **Total Functions**: Number of functions analyzed
- **Files Analyzed**: Number of source files processed

#### Function-Level Breakdown:
- Functions sorted by complexity (highest first)
- Shows file name, line number, function name, and complexity score
- Color-coded complexity levels:
  - **Low** (1-5): Green - Simple, maintainable code
  - **Moderate** (6-10): Yellow - Acceptable complexity
  - **High** (11-20): Orange - Consider refactoring
  - **Very High** (21+): Red - Immediate refactoring needed

### 3. Visual Enhancements

#### Status Indicators:
- ✅ **Healthy**: All checks passed
- ⚠️ **Warning**: Issues found but not critical
- ❌ **Critical**: Serious issues requiring attention

#### Health Score Bars:
- Visual progress bars showing health percentage
- Color-coded based on score (Green 80%+, Yellow 60%+, Red <60%)

#### Professional Layout:
- Unicode box drawing for clean borders
- Consistent spacing and alignment
- Hierarchical information display

### 4. Improved Information Density
- Compact yet comprehensive information display
- Critical metrics prominently featured
- Detailed breakdowns available in verbose mode
- Issue summaries with severity indicators

## Sample Output Structure

```
=== Repository Health Reports ===
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Repository: project-name                                                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Path: /path/to/project                                                          │
│ Language: python                                                                │
│ Status: ⚠️ Warning                                                               │
│ Health Score: 85/100 [█████████████████░░░] 85.0%                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Checks: 7 total, 6 passed, 1 failed, 3 issues found                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Cyclomatic Complexity Analysis                                                 │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Total Complexity: 156                                                          │
│ Maximum Complexity: 23 (Very High)                                             │
│ Average Complexity: 7.8 (Moderate)                                             │
│ Total Functions: 20                                                            │
│ Files Analyzed: 8                                                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Function Complexity Breakdown                                                  │
├─────────────────────────────────────────────────────────────────────────────────┤
│ main.py:145 - process_data() Complexity: 23 (Very High)                        │
│ utils.py:78 - validate_input() Complexity: 15 (High)                           │
│ parser.py:92 - parse_config() Complexity: 12 (High)                            │
│ ... and 17 more functions                                                      │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Health Check Results                                                           │
├─────────────────────────────────────────────────────────────────────────────────┤
│ ✅ Security Scanner (security): Score 90/100                                   │
│ ⚠️ License Compliance (compliance): Score 80/100                               │
│   Issues: 1 found                                                             │
│     • high: No license file found                                             │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Benefits

1. **Better Observability**: Clear view of each repository's health status
2. **Actionable Insights**: Cyclomatic complexity helps identify refactoring candidates
3. **Professional Presentation**: Clean, organized output suitable for reports
4. **Scalable**: Works well for single repos or large multi-repo analyses
5. **Developer Friendly**: Easy to scan and understand at a glance

## Usage

```bash
# Basic health check with new format
repos health

# Verbose mode with function-level complexity details
repos health --verbose

# Target specific repositories
repos health path/to/repo1 path/to/repo2 --verbose
```
