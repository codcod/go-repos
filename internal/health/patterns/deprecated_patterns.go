package patterns

// DeprecatedPattern represents a deprecated component usage pattern
type DeprecatedPattern struct {
	Pattern     string `yaml:"pattern"`
	Replacement string `yaml:"replacement"`
	Severity    string `yaml:"severity"`
	Description string `yaml:"description"`
	Since       string `yaml:"since,omitempty"`
}

// LanguagePatterns holds deprecated patterns for a specific language
type LanguagePatterns struct {
	Language string                       `yaml:"language"`
	Patterns map[string]DeprecatedPattern `yaml:"patterns"`
}

// GetGoDeprecatedPatterns returns Go-specific deprecated patterns
func GetGoDeprecatedPatterns() map[string]DeprecatedPattern {
	return map[string]DeprecatedPattern{
		"ioutil.ReadFile":          {Pattern: "ioutil.ReadFile", Replacement: "os.ReadFile", Severity: "warning", Description: "ioutil.ReadFile deprecated since Go 1.16"},
		"ioutil.WriteFile":         {Pattern: "ioutil.WriteFile", Replacement: "os.WriteFile", Severity: "warning", Description: "ioutil.WriteFile deprecated since Go 1.16"},
		"ioutil.ReadAll":           {Pattern: "ioutil.ReadAll", Replacement: "io.ReadAll", Severity: "warning", Description: "ioutil.ReadAll deprecated since Go 1.16"},
		"ioutil.ReadDir":           {Pattern: "ioutil.ReadDir", Replacement: "os.ReadDir", Severity: "warning", Description: "ioutil.ReadDir deprecated since Go 1.16"},
		"ioutil.TempDir":           {Pattern: "ioutil.TempDir", Replacement: "os.MkdirTemp", Severity: "warning", Description: "ioutil.TempDir deprecated since Go 1.17"},
		"ioutil.TempFile":          {Pattern: "ioutil.TempFile", Replacement: "os.CreateTemp", Severity: "warning", Description: "ioutil.TempFile deprecated since Go 1.17"},
		"golang.org/x/net/context": {Pattern: "golang.org/x/net/context", Replacement: "context", Severity: "critical", Description: "Use standard library context package instead"},
	}
}

// GetJavaDeprecatedPatterns returns Java-specific deprecated patterns
func GetJavaDeprecatedPatterns() map[string]DeprecatedPattern {
	return map[string]DeprecatedPattern{
		"@Deprecated":         {Pattern: "@Deprecated", Replacement: "Check for modern alternative", Severity: "warning", Description: "Using deprecated Java API"},
		"java.util.Date":      {Pattern: "java.util.Date", Replacement: "java.time.LocalDate/LocalDateTime", Severity: "warning", Description: "Use modern Java time API"},
		"SimpleDateFormat":    {Pattern: "SimpleDateFormat", Replacement: "DateTimeFormatter", Severity: "warning", Description: "Use thread-safe DateTimeFormatter"},
		"StringBuffer":        {Pattern: "StringBuffer", Replacement: "StringBuilder", Severity: "warning", Description: "StringBuilder is more efficient for single-threaded use"},
		"java.util.Vector":    {Pattern: "java.util.Vector", Replacement: "java.util.ArrayList", Severity: "warning", Description: "Vector is legacy, use ArrayList or Collections.synchronizedList"},
		"java.util.Hashtable": {Pattern: "java.util.Hashtable", Replacement: "java.util.HashMap", Severity: "warning", Description: "Hashtable is legacy, use HashMap or ConcurrentHashMap"},
		"java.util.Stack":     {Pattern: "java.util.Stack", Replacement: "java.util.Deque", Severity: "warning", Description: "Stack is legacy, use ArrayDeque"},
	}
}

// GetJavaScriptDeprecatedPatterns returns JavaScript/TypeScript-specific deprecated patterns
func GetJavaScriptDeprecatedPatterns() map[string]DeprecatedPattern {
	return map[string]DeprecatedPattern{
		"var ":                             {Pattern: "var ", Replacement: "const/let", Severity: "warning", Description: "Use const or let instead of var"},
		"$.ajax":                           {Pattern: "$.ajax", Replacement: "fetch() or axios", Severity: "warning", Description: "Use modern HTTP client instead of jQuery ajax"},
		"componentWillMount":               {Pattern: "componentWillMount", Replacement: "componentDidMount", Severity: "critical", Description: "React lifecycle method deprecated"},
		"componentWillReceiveProps":        {Pattern: "componentWillReceiveProps", Replacement: "componentDidUpdate", Severity: "critical", Description: "React lifecycle method deprecated"},
		"componentWillUpdate":              {Pattern: "componentWillUpdate", Replacement: "componentDidUpdate", Severity: "critical", Description: "React lifecycle method deprecated"},
		"UNSAFE_componentWillMount":        {Pattern: "UNSAFE_componentWillMount", Replacement: "componentDidMount", Severity: "warning", Description: "Unsafe React lifecycle method"},
		"UNSAFE_componentWillReceiveProps": {Pattern: "UNSAFE_componentWillReceiveProps", Replacement: "componentDidUpdate", Severity: "warning", Description: "Unsafe React lifecycle method"},
		"UNSAFE_componentWillUpdate":       {Pattern: "UNSAFE_componentWillUpdate", Replacement: "componentDidUpdate", Severity: "warning", Description: "Unsafe React lifecycle method"},
		"ReactDOM.findDOMNode":             {Pattern: "ReactDOM.findDOMNode", Replacement: "useRef hook", Severity: "warning", Description: "findDOMNode is deprecated in StrictMode"},
		"String.prototype.substr":          {Pattern: ".substr(", Replacement: ".substring(", Severity: "warning", Description: "substr() is deprecated, use substring()"},
	}
}

// GetPythonDeprecatedPatterns returns Python-specific deprecated patterns
func GetPythonDeprecatedPatterns() map[string]DeprecatedPattern {
	return map[string]DeprecatedPattern{
		"imp.":                       {Pattern: "imp.", Replacement: "importlib", Severity: "warning", Description: "imp module is deprecated since Python 3.4"},
		"optparse":                   {Pattern: "optparse", Replacement: "argparse", Severity: "warning", Description: "optparse is deprecated since Python 2.7"},
		"platform.dist":              {Pattern: "platform.dist", Replacement: "platform.freedesktop_os_release", Severity: "warning", Description: "platform.dist deprecated since Python 3.5"},
		"cgi.escape":                 {Pattern: "cgi.escape", Replacement: "html.escape", Severity: "warning", Description: "cgi.escape deprecated since Python 3.2"},
		"collections.Mapping":        {Pattern: "collections.Mapping", Replacement: "collections.abc.Mapping", Severity: "warning", Description: "Import from collections.abc instead"},
		"collections.MutableMapping": {Pattern: "collections.MutableMapping", Replacement: "collections.abc.MutableMapping", Severity: "warning", Description: "Import from collections.abc instead"},
		"collections.Sequence":       {Pattern: "collections.Sequence", Replacement: "collections.abc.Sequence", Severity: "warning", Description: "Import from collections.abc instead"},
		"collections.Iterable":       {Pattern: "collections.Iterable", Replacement: "collections.abc.Iterable", Severity: "warning", Description: "Import from collections.abc instead"},
		"asyncio.coroutine":          {Pattern: "asyncio.coroutine", Replacement: "async def", Severity: "warning", Description: "Use async/await syntax instead of @asyncio.coroutine"},
	}
}

// GetDockerDeprecatedPatterns returns Docker-specific deprecated patterns
func GetDockerDeprecatedPatterns() map[string]DeprecatedPattern {
	return map[string]DeprecatedPattern{
		"MAINTAINER":        {Pattern: "MAINTAINER", Replacement: "LABEL maintainer=", Severity: "warning", Description: "MAINTAINER instruction is deprecated"},
		"FROM ubuntu:14.04": {Pattern: "FROM ubuntu:14.04", Replacement: "FROM ubuntu:20.04 or later", Severity: "critical", Description: "Ubuntu 14.04 is end-of-life"},
		"FROM ubuntu:16.04": {Pattern: "FROM ubuntu:16.04", Replacement: "FROM ubuntu:20.04 or later", Severity: "warning", Description: "Ubuntu 16.04 is end-of-life"},
		"FROM centos:6":     {Pattern: "FROM centos:6", Replacement: "FROM centos:8 or rocky/alma", Severity: "critical", Description: "CentOS 6 is end-of-life"},
		"FROM centos:7":     {Pattern: "FROM centos:7", Replacement: "FROM centos:8 or rocky/alma", Severity: "warning", Description: "CentOS 7 will be end-of-life soon"},
		"FROM node:10":      {Pattern: "FROM node:10", Replacement: "FROM node:18 or later", Severity: "critical", Description: "Node.js 10 is end-of-life"},
		"FROM node:12":      {Pattern: "FROM node:12", Replacement: "FROM node:18 or later", Severity: "warning", Description: "Node.js 12 is end-of-life"},
		"FROM python:2":     {Pattern: "FROM python:2", Replacement: "FROM python:3", Severity: "critical", Description: "Python 2 is end-of-life"},
	}
}

// GetKubernetesDeprecatedPatterns returns Kubernetes-specific deprecated patterns
func GetKubernetesDeprecatedPatterns() map[string]DeprecatedPattern {
	return map[string]DeprecatedPattern{
		"apiVersion: extensions/v1beta1":                {Pattern: "apiVersion: extensions/v1beta1", Replacement: "apps/v1", Severity: "critical", Description: "extensions/v1beta1 API is deprecated"},
		"apiVersion: apps/v1beta1":                      {Pattern: "apiVersion: apps/v1beta1", Replacement: "apps/v1", Severity: "critical", Description: "apps/v1beta1 API is deprecated"},
		"apiVersion: apps/v1beta2":                      {Pattern: "apiVersion: apps/v1beta2", Replacement: "apps/v1", Severity: "warning", Description: "apps/v1beta2 API is deprecated"},
		"apiVersion: networking.k8s.io/v1beta1":         {Pattern: "apiVersion: networking.k8s.io/v1beta1", Replacement: "networking.k8s.io/v1", Severity: "warning", Description: "networking.k8s.io/v1beta1 API is deprecated"},
		"apiVersion: policy/v1beta1":                    {Pattern: "apiVersion: policy/v1beta1", Replacement: "policy/v1", Severity: "warning", Description: "policy/v1beta1 API is deprecated"},
		"apiVersion: rbac.authorization.k8s.io/v1beta1": {Pattern: "apiVersion: rbac.authorization.k8s.io/v1beta1", Replacement: "rbac.authorization.k8s.io/v1", Severity: "warning", Description: "rbac.authorization.k8s.io/v1beta1 API is deprecated"},
	}
}
