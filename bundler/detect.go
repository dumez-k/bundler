package bundler

import (
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

//go:generate faux --interface VersionParser --output fakes/version_parser.go
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

type BuildPlanMetadata struct {
	VersionSource string `toml:"version-source"`
}

func Detect(buildpackYMLParser, gemfileLockParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requirements []packit.BuildPlanRequirement

		version, err := buildpackYMLParser.ParseVersion(filepath.Join(context.WorkingDir, BuildpackYMLSource))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name:    Bundler,
				Version: version,
				Metadata: BuildPlanMetadata{
					VersionSource: BuildpackYMLSource,
				},
			})
		}

		version, err = gemfileLockParser.ParseVersion(filepath.Join(context.WorkingDir, GemfileLockSource))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name:    Bundler,
				Version: version,
				Metadata: BuildPlanMetadata{
					VersionSource: GemfileLockSource,
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: Bundler},
				},
				Requires: requirements,
			},
		}, nil
	}
}
