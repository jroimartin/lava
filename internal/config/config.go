// Copyright 2023 Adevinta

// Package config implements parsing of Lava configurations.
package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	agentconfig "github.com/adevinta/vulcan-agent/config"
	types "github.com/adevinta/vulcan-types"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

var (
	// ErrInvalidLavaVersion means that the Lava version does not
	// have a valid format according to the Semantic Versioning
	// Specification.
	ErrInvalidLavaVersion = errors.New("invalid Lava version")

	// ErrNoTargets means that no targets were specified.
	ErrNoTargets = errors.New("no targets")

	// ErrNoTargetIdentifier means that the target does not have
	// an identifier.
	ErrNoTargetIdentifier = errors.New("no target identifier")

	// ErrInvalidAssetType means that the asset type is invalid.
	ErrInvalidAssetType = errors.New("invalid asset type")

	// ErrInvalidSeverity means that the severity is invalid.
	ErrInvalidSeverity = errors.New("invalid severity")

	// ErrInvalidOutputFormat means that the output format is
	// invalid.
	ErrInvalidOutputFormat = errors.New("invalid output format")
)

// Config represents a Lava configuration.
type Config struct {
	// LavaVersion is the minimum required version of Lava.
	LavaVersion string `yaml:"lava"`

	// AgentConfig is the configuration of the vulcan-agent.
	AgentConfig AgentConfig `yaml:"agent"`

	// ReportConfig is the configuration of the report.
	ReportConfig ReportConfig `yaml:"report"`

	// ChecktypesURLs is a list of URLs pointing to checktypes
	// catalogs.
	ChecktypesURLs []string `yaml:"checktypesURLs"`

	// Targets is the list of targets.
	Targets []Target `yaml:"targets"`

	// LogLevel is the logging level.
	LogLevel slog.Level `yaml:"logLevel"`
}

// Parse returns a parsed Lava configuration given an [io.Reader].
func Parse(r io.Reader) (Config, error) {
	var cfg Config
	if err := yaml.NewDecoder(r).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}
	return cfg, nil
}

// ParseFile returns a parsed Lava configuration given a path to a
// file.
func ParseFile(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config file: %w", err)
	}
	defer f.Close()
	return Parse(f)
}

// validate validates the Lava configuration.
func (c *Config) validate() error {
	// Lava version validation.
	if !semver.IsValid(c.LavaVersion) {
		return ErrInvalidLavaVersion
	}

	// Targets validation.
	if len(c.Targets) == 0 {
		return ErrNoTargets
	}
	for _, target := range c.Targets {
		if target.Identifier == "" {
			return ErrNoTargetIdentifier
		}
	}
	return nil
}

// AgentConfig is the configuration passed to the vulcan-agent.
type AgentConfig struct {
	// PullPolicy is the pull policy passed to vulcan-agent.
	PullPolicy agentconfig.PullPolicy `yaml:"pullPolicy"`

	// Parallel is the maximum number of checks that can run in
	// parallel.
	Parallel int `yaml:"parallel"`

	// Vars is the environment variables required by the Vulcan
	// checktypes.
	Vars map[string]string `yaml:"vars"`

	// RegistriesAuth contains the credentials for a set of
	// container registries.
	RegistriesAuth []RegistryAuth `yaml:"registriesAuth"`
}

// ReportConfig is the configuration of the report.
type ReportConfig struct {
	// Severity is the minimum severity required to report a
	// finding.
	Severity Severity `yaml:"severity"`

	// Format is the output format.
	Format OutputFormat `yaml:"format"`

	// OutputFile is the path of the output file.
	OutputFile string `yaml:"outputFile"`

	// Exclusions is a list of findings that will be ignored. For
	// instance, accepted risks, false positives, etc.
	Exclusions []Exclusion `yaml:"exclusions"`
}

// Target represents the target of a scan.
type Target struct {
	// Identifier is a string that identifies the target. For
	// instance, a path, a URL, a Docker image, etc.
	Identifier string `yaml:"identifier"`

	// AssetType is the asset type of the target.
	AssetType AssetType `yaml:"assetType"`

	// Options is a list of specific options for the target.
	Options map[string]any `yaml:"options"`
}

// RegistryAuth contains the credentials for a container registry.
type RegistryAuth struct {
	// Server is the URL of the registry.
	Server string `yaml:"server"`

	// Username is the username used to log into the registry.
	Username string `yaml:"username"`

	// Password is the password used to log into the registry.
	Password string `yaml:"password"`
}

// Severity is the severity of a given finding.
type Severity int

// Severity levels.
const (
	SeverityCritical Severity = 1
	SeverityHigh     Severity = 0
	SeverityMedium   Severity = -1
	SeverityLow      Severity = -2
	SeverityInfo     Severity = -3
)

var severityNames = map[string]Severity{
	"critical": SeverityCritical,
	"high":     SeverityHigh,
	"medium":   SeverityMedium,
	"low":      SeverityLow,
	"info":     SeverityInfo,
}

// parseSeverity converts a string into a [Severity] value.
func parseSeverity(severity string) (Severity, error) {
	if val, ok := severityNames[severity]; ok {
		return val, nil
	}

	var zero Severity
	return zero, fmt.Errorf("%w: %v", ErrInvalidSeverity, severity)
}

// UnmarshalYAML decodes a Severity yaml node containing a string into
// a [Severity] value. It returns error if the provided string does
// not match any known severity.
func (s *Severity) UnmarshalYAML(value *yaml.Node) error {
	severity, err := parseSeverity(value.Value)
	if err != nil {
		return err
	}
	*s = severity
	return nil
}

// OutputFormat is the format of the generated report.
type OutputFormat int

// Output formats available for the report.
const (
	OutputFormatJSON OutputFormat = 0
)

var outputFormatNames = map[string]OutputFormat{
	"json": OutputFormatJSON,
}

// parseOutputFormat converts a string into an [OutputFormat] value.
func parseOutputFormat(format string) (OutputFormat, error) {
	if val, ok := outputFormatNames[strings.ToLower(format)]; ok {
		return val, nil
	}

	var zero OutputFormat
	return zero, fmt.Errorf("%w: %v", ErrInvalidOutputFormat, format)
}

// UnmarshalYAML decodes an OutputFormat yaml node containing a string
// into an [OutputFormat] value. It returns error if the provided
// string does not match any known output format.
func (f *OutputFormat) UnmarshalYAML(value *yaml.Node) error {
	format, err := parseOutputFormat(value.Value)
	if err != nil {
		return err
	}
	*f = format
	return nil
}

// Exclusion represents the criteria to exclude a given finding.
type Exclusion struct {
	// Target is the name of the affected target.
	Target string `yaml:"target"`

	// Resource is the name of the affected resource.
	Resource string `yaml:"resource"`

	// Fingerprint defines the context in where the vulnerability
	// has been found. It includes the checktype image, the
	// affected target, the asset type and the checktype options.
	Fingerprint string `yaml:"fingerprint"`

	// Summary is a short description of the exclusion.
	Summary string `yaml:"summary"`

	// Description describes the exclusion.
	Description string `yaml:"description"`
}

// AssetType represents the type of an asset.
type AssetType types.AssetType

// UnmarshalYAML decodes an AssetType yaml node containing a string
// into an [AssetType] value. It returns error if the provided string
// does not match any known asset type.
func (t *AssetType) UnmarshalYAML(value *yaml.Node) error {
	at, err := types.Parse(value.Value)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAssetType, value.Value)
	}
	*t = AssetType(at)
	return nil
}