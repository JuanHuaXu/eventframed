package claimsexperiment

import (
	"context"
	"fmt"
)

const (
	RunRoleExploratory  = "exploratory"
	RunRoleDesign       = "design"
	RunRoleConfirmation = "confirmation"
)

type RunOptions struct {
	SeedBase       int64
	Role           string
	DesignSeedBase *int64
}

func RunSuite(ctx context.Context) (SuiteReport, error) {
	return RunSuiteWithOptions(ctx, RunOptions{SeedBase: 982451653, Role: RunRoleExploratory})
}

func RunSuiteWithSeed(ctx context.Context, seedBase int64) (SuiteReport, error) {
	return RunSuiteWithOptions(ctx, RunOptions{SeedBase: seedBase, Role: RunRoleExploratory})
}

func RunSuiteWithOptions(ctx context.Context, options RunOptions) (SuiteReport, error) {
	if err := validateRunOptions(options); err != nil {
		return SuiteReport{}, err
	}
	residualReport, err := RunResidualUtility(ctx)
	if err != nil {
		return SuiteReport{}, err
	}
	heterogeneousResidual := RunHeterogeneousResidual(options.SeedBase + 1000003)
	antiPigeonReport, err := RunAntiPigeonGranularity(ctx)
	if err != nil {
		return SuiteReport{}, err
	}
	groupControl, err := RunGroupAuthorityControl(ctx)
	if err != nil {
		return SuiteReport{}, err
	}
	manifest := RunManifest{Role: options.Role, SeedBase: options.SeedBase, DesignSeedBase: options.DesignSeedBase, ProtocolFrozenBeforeRun: true}
	if options.Role == RunRoleConfirmation {
		distinct := options.DesignSeedBase != nil && *options.DesignSeedBase != options.SeedBase
		manifest.ConfirmationSeedDistinct = &distinct
	}
	return SuiteReport{
		SchemaVersion:           SchemaVersion,
		Protocol:                FrozenProtocol(),
		Run:                     manifest,
		Residual:                residualReport,
		ResidualHeterogeneous:   heterogeneousResidual,
		ResidualSafeReplacement: RunSafeResidualReplacement(options.SeedBase + 3000017),
		OmittedAudit:            RunOmittedAuditCoverage(options.SeedBase + 4000037),
		AntiPigeon:              antiPigeonReport,
		GroupComparison:         RunGroupComparison(options.SeedBase + 2000003),
		Changepoint:             RunChangepointAdaptationWithSeed(options.SeedBase),
		IntegrationControls:     IntegrationControlsReport{GroupAuthority: groupControl},
	}, nil
}

func validateRunOptions(options RunOptions) error {
	switch options.Role {
	case RunRoleExploratory, RunRoleDesign:
		if options.DesignSeedBase != nil {
			return fmt.Errorf("design seed base is only valid for confirmation runs")
		}
	case RunRoleConfirmation:
		if options.DesignSeedBase == nil {
			return fmt.Errorf("confirmation runs require the corresponding design seed base")
		}
		if *options.DesignSeedBase == options.SeedBase {
			return fmt.Errorf("confirmation seed base must differ from design seed base")
		}
	default:
		return fmt.Errorf("invalid run role %q", options.Role)
	}
	return nil
}
